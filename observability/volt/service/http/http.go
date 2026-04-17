package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.ngrok.com/ngrok/v2"
)

const (
	name            = "volt.service"
	maxBodyLen      = 1 << 20 // 1 MiB
	shutdownTimeout = 30 * time.Second
)

func randMs(min, max int) time.Duration {
	i := min + rand.IntN(max-min)
	return time.Duration(i) * time.Millisecond
}

// responseWriter wraps http.ResponseWriter to capture the written status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (r *responseWriter) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// httpMetrics holds Prometheus metrics for the server.
type httpMetrics struct {
	http.Handler

	reqGauge    *prometheus.GaugeVec
	reqCounter  *prometheus.CounterVec
	reqDuration *prometheus.HistogramVec
}

func newHTTPMetrics() *httpMetrics {
	reqGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_active",
			Help: "The number of in-flight HTTP requests",
		},
		[]string{"name", "req_method", "req_route"},
	)

	reqCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "The total number of HTTP requests",
		},
		[]string{"name", "req_method", "req_route", "resp_status"},
	)

	reqDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_requests_duration_seconds",
			Help:    "The duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"name", "req_method", "req_route", "resp_status"},
	)

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		reqGauge,
		reqCounter,
		reqDuration,
	)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry: reg,
	})

	return &httpMetrics{
		Handler:     handler,
		reqGauge:    reqGauge,
		reqCounter:  reqCounter,
		reqDuration: reqDuration,
	}
}

// service represents a generic HTTP service.
type service struct {
	mux     *http.ServeMux
	metrics *httpMetrics
	logger  *slog.Logger
}

func newService() *service {
	mux := http.NewServeMux()
	metrics := newHTTPMetrics()

	logger := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	mux.Handle("/metrics", metrics)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &service{
		mux:     mux,
		metrics: metrics,
		logger:  logger,
	}
}

func (s *service) withInstrumentation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := r.URL.Path
		req_uuid := uuid.NewString()

		s.metrics.reqGauge.WithLabelValues(name, r.Method, route).Inc()
		defer s.metrics.reqGauge.WithLabelValues(name, r.Method, route).Dec()

		rw := &responseWriter{ResponseWriter: w}
		next(rw, r)

		duration := time.Since(start)
		status := strconv.Itoa(rw.status)

		s.metrics.reqCounter.WithLabelValues(name, r.Method, route, status).Inc()
		s.metrics.reqDuration.WithLabelValues(name, r.Method, route, status).Observe(duration.Seconds())

		s.logger.Info("[http] request handled.",
			slog.String("name", name),
			slog.String("req_uuid", req_uuid),
			slog.String("req_method", r.Method),
			slog.String("req_path", r.URL.Path),
			slog.String("req_route", route),
			slog.Int("resp_status", rw.status),
			slog.Duration("duration", duration),
		)
	}
}

func (s *service) RegisterRoute(route string, handler http.HandlerFunc) {
	s.mux.HandleFunc(route, s.withInstrumentation(handler))
}

func (s *service) Start(port int, ngrokEnabled bool, ngrokAuthToken string) {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	tn := &tunnel{
		logger: s.logger,
	}

	// Local server
	go func() {
		s.logger.Info("[http] starting server ...", slog.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("[http] server stopped unexpectedly.", "error", err)
			os.Exit(1)
		}
	}()

	// ngrok tunnel
	if ngrokEnabled {
		go func() {
			s.logger.Info("[http] opening ngrok tunnel ...")
			if err := tn.Open(port, ngrokAuthToken); err != nil {
				s.logger.Error("[http] error opening ngrok tunnel.", "error", err)
				os.Exit(1)
			}
		}()
	}

	// Gracefully shutdown the server on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	s.logger.Info("[http] shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Closing the ngrok tunnel
	if err := tn.Close(ctx); err != nil {
		s.logger.Error("[http] error closing ngrok tunnel.", "error", err)
	}

	// Closing the local server
	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error("[http] error shutting down server.", "error", err)
	}

	s.logger.Info("[http] server shutdown gracefully.")
}

type tunnel struct {
	logger    *slog.Logger
	forwarder ngrok.EndpointForwarder
}

func (t *tunnel) Open(port int, authToken string) error {
	ctx := context.Background()
	addr := fmt.Sprintf("http://localhost:%d", port)

	agent, err := ngrok.NewAgent(
		ngrok.WithAuthtoken(authToken),
	)

	if err != nil {
		return err
	}

	t.forwarder, err = agent.Forward(ctx,
		ngrok.WithUpstream(addr),
		ngrok.WithName("Volt Webhook Service"),
		ngrok.WithDescription("Exposing the Volt webhook service to the public internet."),
	)

	if err != nil {
		return err
	}

	t.logger.Info("[http] ngrok tunnel established",
		slog.String("from", t.forwarder.URL().String()),
		slog.String("to", addr),
	)

	return nil
}

func (t *tunnel) Close(ctx context.Context) error {
	if t.forwarder != nil {
		return t.forwarder.CloseWithContext(ctx)
	}

	return nil
}
