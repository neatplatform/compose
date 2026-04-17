package service

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metrics holds Prometheus metrics for the server.
type metrics struct {
	http.Handler

	reqGauge    *prometheus.GaugeVec
	reqCounter  *prometheus.CounterVec
	reqDuration *prometheus.HistogramVec
}

func newMetrics() *metrics {
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

	return &metrics{
		Handler:     handler,
		reqGauge:    reqGauge,
		reqCounter:  reqCounter,
		reqDuration: reqDuration,
	}
}
