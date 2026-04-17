package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/neatplatform/compose/observability/slack/client"
	"github.com/neatplatform/compose/observability/slack/schema"
)

const (
	name            = "slack-service"
	maxBodyLen      = 1 << 20 // 1 MiB
	shutdownTimeout = 30 * time.Second

	slackTimestampHeader  = "X-Slack-Request-Timestamp"
	slackSignatureHeader  = "X-Slack-Signature"
	maxTimestampAgeInSecs = 300 // 5 minutes
	signingVersion        = "v0"

	baseURL = "https://slack.com/api"
)

var (
	locationOptionGroups = []schema.OptionGroup{
		{
			Label: schema.Text{
				Type: "plain_text",
				Text: "CANADA",
			},
			Options: []schema.Option{
				{
					Value: "toronto",
					Text: schema.Text{
						Type: "plain_text",
						Text: "Toronto",
					},
				},
				{
					Value: "vancouver",
					Text: schema.Text{
						Type: "plain_text",
						Text: "Vancouver",
					},
				},
			},
		},
		{
			Label: schema.Text{
				Type: "plain_text",
				Text: "ENGLAND",
			},
			Options: []schema.Option{
				{
					Value: "london",
					Text: schema.Text{
						Type: "plain_text",
						Text: "London",
					},
				},
				{
					Value: "liverpool",
					Text: schema.Text{
						Type: "plain_text",
						Text: "Liverpool",
					},
				},
			},
		},
		{
			Label: schema.Text{
				Type: "plain_text",
				Text: "ITALY",
			},
			Options: []schema.Option{
				{
					Value: "rome",
					Text: schema.Text{
						Type: "plain_text",
						Text: "Rome",
					},
				},
				{
					Value: "milan",
					Text: schema.Text{
						Type: "plain_text",
						Text: "Milan",
					},
				},
			},
		},
	}
)

type Service struct {
	mux           *http.ServeMux
	metrics       *metrics
	logger        *slog.Logger
	signingSecret string
	client        *client.Client
}

func New(logger *slog.Logger, signingSecret string, client *client.Client) *Service {
	mux := http.NewServeMux()
	metrics := newMetrics()

	mux.Handle("/metrics", metrics)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s := &Service{
		mux:           mux,
		metrics:       metrics,
		logger:        logger,
		signingSecret: signingSecret,
		client:        client,
	}

	s.mux.HandleFunc("/commands", s.withInstrumentation(s.prepare(s.verify(s.commands))))
	s.mux.HandleFunc("/events", s.withInstrumentation(s.prepare(s.verify(s.events))))
	s.mux.HandleFunc("/interactions", s.withInstrumentation(s.prepare(s.verify(s.interactions))))
	s.mux.HandleFunc("/options", s.withInstrumentation(s.prepare(s.verify(s.options))))

	return s
}

func (s *Service) Start(port int, ngrokEnabled bool, ngrokAuthToken string) {
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
		s.logger.Info("Starting server ...", slog.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Server stopped unexpectedly.", "error", err)
			os.Exit(1)
		}
	}()

	// ngrok tunnel
	if ngrokEnabled {
		go func() {
			s.logger.Info("Opening ngrok tunnel ...")
			if err := tn.Open(port, ngrokAuthToken); err != nil {
				s.logger.Error("Error opening ngrok tunnel.", "error", err)
				os.Exit(1)
			}
		}()
	}

	// Gracefully shutdown the server on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	s.logger.Info("Shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Closing the ngrok tunnel
	if err := tn.Close(ctx); err != nil {
		s.logger.Error("Error closing ngrok tunnel.", "error", err)
	}

	// Closing the local server
	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error("Error shutting down server.", "error", err)
	}

	s.logger.Info("Server shutdown gracefully.")
}

func (s *Service) withInstrumentation(next http.HandlerFunc) http.HandlerFunc {
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

		s.logger.Info("Request handled.",
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

func (s *Service) prepare(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyLen))
		if err != nil {
			s.logger.Error("[prepare] Error reading request body.", "error", err)
			http.Error(w, "Failed to read request body.", http.StatusInternalServerError)
			return
		}

		// Restore the original request body for the next handler.
		r.Body = io.NopCloser(bytes.NewReader(body))

		// Add the raw body to the request context for the next handler.
		r = r.WithContext(contextWithBody(r.Context(), body))

		var bodyVal slog.Value

		switch r.Header.Get("Content-Type") {
		case "application/x-www-form-urlencoded":
			formBody, err := url.ParseQuery(string(body))
			if err != nil {
				s.logger.Error("[prepare] Error parsing request body.", "error", err)
				http.Error(w, "Failed to parse request body.", http.StatusBadRequest)
				return
			}

			bodyVal = slog.AnyValue(formBody)

		case "application/json":
			var jsonBody map[string]any
			if err := json.Unmarshal(body, &jsonBody); err != nil {
				s.logger.Error("[prepare] Error parsing request body.", "error", err)
				http.Error(w, "Failed to parse request body.", http.StatusBadRequest)
				return
			}

			bodyVal = slog.AnyValue(jsonBody)

		default:
			bodyVal = slog.StringValue(string(body))
		}

		s.logger.Debug("[prepare] New request.",
			slog.Any("headers", r.Header),
			slog.Any("body", bodyVal),
		)

		// Proceed to the next handler
		next(w, r)
	}
}

func (s *Service) verify(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Verify request timestamp
		timestamp := r.Header.Get(slackTimestampHeader)

		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			s.logger.Error("[verify] Error parsing request timestamp.", "error", err)
			http.Error(w, "Failed to parse Slack timestamp.", http.StatusBadRequest)
			return
		}

		age := time.Now().Unix() - ts

		// Workaround for clock skew.
		if age < 0 {
			age = 0
		}

		if age > maxTimestampAgeInSecs {
			s.logger.Error("[verify] Request timestamp exceeds maximum age.",
				slog.String("slack_request_timestamp", timestamp),
				slog.Int64("slack_request_age", age),
				slog.Int("slack_request_max_age", maxTimestampAgeInSecs),
			)
			http.Error(w, "Slack timestamp exceeds maximum age.", http.StatusBadRequest)
			return
		}

		// Verify request signature
		signature := r.Header.Get(slackSignatureHeader)

		// Assuming the body has been consumed and saved by the prepare middleware.
		body, _ := bodyFromContext(r.Context())

		base := fmt.Sprintf("%s:%s:%s", signingVersion, timestamp, body)
		mac := hmac.New(sha256.New, []byte(s.signingSecret))
		mac.Write([]byte(base))
		expected := fmt.Sprintf("%s=%x", signingVersion, mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expected)) {
			s.logger.Error("[verify] Invalid Slack signature.",
				slog.String("signature", signature),
				slog.String("expected_signature", expected),
			)
			http.Error(w, "Slack signature is invalid.", http.StatusBadRequest)
			return
		}

		// Verify request method
		if r.Method != http.MethodPost {
			s.logger.Error("[verify] Unexpected HTTP method.", slog.String("method", r.Method))
			http.Error(w, "Only POST is allowed.", http.StatusMethodNotAllowed)
			return
		}

		s.logger.Info("[verify] Request verified.")

		// Proceed to the next handler
		next(w, r)
	}
}

func (s *Service) commands(w http.ResponseWriter, r *http.Request) {
	// Assuming the body has been consumed and saved by the prepare middleware.
	body, _ := bodyFromContext(r.Context())

	vals, err := url.ParseQuery(string(body))
	if err != nil {
		s.logger.Error("[command] Error parsing command.", "error", err)
		http.Error(w, "Failed to parse slash command.", http.StatusBadRequest)
		return
	}

	cmd := schema.Command{
		TeamID:      vals.Get("team_id"),
		TeamDomain:  vals.Get("team_domain"),
		ChannelID:   vals.Get("channel_id"),
		ChannelName: vals.Get("channel_name"),
		UserID:      vals.Get("user_id"),
		UserName:    vals.Get("user_name"),
		Command:     vals.Get("command"),
		Text:        vals.Get("text"),
		ResponseURL: vals.Get("response_url"),
		TriggerID:   vals.Get("trigger_id"),
		APIAppID:    vals.Get("api_app_id"),
	}

	// Acknowledge the request within 3000 milliseconds.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schema.MessageResponse{
		ResponseType: "in_channel",
		Text:         "Received!",
	})

	s.logger.Info("[command] Acknowledged!",
		slog.String("channel", cmd.ChannelName),
		slog.String("user", cmd.UserName),
		slog.String("command", cmd.Command),
		slog.String("text", cmd.Text),
	)

	go s.handleCommand(cmd)
}

func (s *Service) handleCommand(cmd schema.Command) {
	first := schema.MessageResponse{
		ResponseType: "in_channel",
		Text:         fmt.Sprintf("Hello, %s!", cmd.Text),
	}

	if _, err := s.client.PostMessageResponse(cmd.ResponseURL, first); err != nil {
		s.logger.Error("[command] Error posting message response.", "error", err)
	}

	second := schema.MessageResponse{
		ResponseType: "in_channel",
		Text:         fmt.Sprintf("Goodbye, %s!", cmd.Text),
	}

	if _, err := s.client.PostMessageResponse(cmd.ResponseURL, second); err != nil {
		s.logger.Error("[command] Error posting message response.", "error", err)
	}

	s.logger.Info("[command] Handled!",
		slog.String("channel", cmd.ChannelName),
		slog.String("user", cmd.UserName),
		slog.String("command", cmd.Command),
		slog.String("text", cmd.Text),
	)
}

func (s *Service) events(w http.ResponseWriter, r *http.Request) {
	// Assuming the body has been consumed and saved by the prepare middleware.
	body, _ := bodyFromContext(r.Context())

	var env schema.EventEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		s.logger.Error("[event] Error parsing event callback request.", "error", err)
		http.Error(w, "Failed to parse event callback request.", http.StatusBadRequest)
		return
	}

	switch env.Type {
	case "url_verification":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"challenge": env.Challenge,
		})

		s.logger.Info("[event] Handled URL verification.")

	case "event_callback":
		if env.Event == nil {
			s.logger.Error("[event] Missing inner event.")
			http.Error(w, "Missing inner event.", http.StatusBadRequest)
			return
		}

		// Acknowledge the request within 3000 milliseconds.
		w.WriteHeader(http.StatusOK)

		s.logger.Info("[event] Acknowledged!")

		go s.handleEventCallback(env)

	default:
		s.logger.Error("[event] Unsupported type: " + env.Type)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *Service) handleEventCallback(env schema.EventEnvelope) {
	var base schema.BaseEvent
	if err := json.Unmarshal(env.Event, &base); err != nil {
		s.logger.Error("[event] Error parsing event type.", "error", err)
		return
	}

	switch base.Type {
	case "message":
		var e schema.MessageEvent
		if err := json.Unmarshal(env.Event, &e); err != nil {
			s.logger.Error("[event] Error parsing message event.", "error", err)
			return
		}

		req := &schema.AddReactionRequest{
			Channel:   e.Channel,
			Timestamp: e.TS,
			Name:      "thumbsup",
		}

		if _, err := s.client.AddReaction(req); err != nil {
			s.logger.Error("[event] Error adding reaction.", "error", err)
			return
		}

		s.logger.Info("[event] Handled message event.",
			slog.Any("event", e),
		)

	case "app_mention":
		var e schema.AppMentionEvent
		if err := json.Unmarshal(env.Event, &e); err != nil {
			s.logger.Error("[event] Error parsing app_mention event.", "error", err)
			return
		}

		req := &schema.PostMessageRequest{
			Channel:  e.Channel,
			Text:     "Noted!",
			ThreadTS: e.TS, // Reply in thread
		}

		if _, err := s.client.PostMessage(req); err != nil {
			s.logger.Error("[event] Error posting message.", "error", err)
			return
		}

		s.logger.Info("[event] Handled app_mention event.",
			slog.Any("event", e),
		)

	default:
		s.logger.Error("[event] Unsupported event type: " + base.Type)
		return
	}
}

func (s *Service) interactions(w http.ResponseWriter, r *http.Request) {
	// Assuming the body has been consumed and saved by the prepare middleware.
	body, _ := bodyFromContext(r.Context())

	values, err := url.ParseQuery(string(body))
	if err != nil {
		s.logger.Error("[interaction] Error parsing interaction request.", "error", err)
		http.Error(w, "Failed to parse interaction request.", http.StatusBadRequest)
		return
	}

	rawPayload := values.Get("payload")
	if rawPayload == "" {
		s.logger.Error("[interaction] Missing interaction payload.", "error", err)
		http.Error(w, "Missing interaction payload.", http.StatusBadRequest)
		return
	}

	// Acknowledge the request within 3000 milliseconds.
	w.WriteHeader(http.StatusOK)

	s.logger.Info("[interaction] Acknowledged!")

	go s.handleInteraction(rawPayload)
}

func (s *Service) handleInteraction(rawPayload string) {
	var base schema.InteractionPayloadBase
	if err := json.Unmarshal([]byte(rawPayload), &base); err != nil {
		s.logger.Error("[interaction] Error parsing interaction payload.", "error", err)
		return
	}

	switch base.Type {
	// https://docs.slack.dev/reference/interaction-payloads/block_actions-payload
	case "block_actions":
		var payload schema.InteractionPayloadBlockActions
		if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
			s.logger.Error("[interaction] Error parsing block_actions interaction payload.", "error", err)
			return
		}

		if len(payload.Actions) == 1 {
			var base schema.ActionBase
			if err := json.Unmarshal([]byte(payload.Actions[0]), &base); err != nil {
				s.logger.Error("[interaction] Error parsing action.", "error", err)
				return
			}

			switch base.Type {
			case "button":
				var action schema.ActionButton
				if err := json.Unmarshal([]byte(payload.Actions[0]), &action); err != nil {
					s.logger.Error("[interaction] Error parsing button action.", "error", err)
					return
				}

				if action.BlockID == "form_block" && action.ActionID == "fill_button" && action.Value == "fill_1234" {
					req := &schema.OpenViewRequest{
						TriggerID: payload.TriggerID,
						View:      []byte(client.ModalFixture),
					}

					if _, err := s.client.OpenView(req); err != nil {
						s.logger.Error("[interaction] Error opening view.", "error", err)
					}
				}
			}
		}

		s.logger.Info("[interaction] Handled block_actions interaction.",
			slog.Any("payload", payload),
		)

	// https://docs.slack.dev/reference/interaction-payloads/shortcuts-interaction-payload
	case "shortcut":
		var payload schema.InteractionPayloadShortcut
		if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
			s.logger.Error("[interaction] Error parsing shortcut interaction payload.", "error", err)
			return
		}

		s.logger.Info("[interaction] Handled shortcut interaction.",
			slog.Any("payload", payload),
		)

	// https://docs.slack.dev/reference/interaction-payloads/shortcuts-interaction-payload
	case "message_action":
		var payload schema.InteractionPayloadShortcut
		if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
			s.logger.Error("[interaction] Error parsing shortcut interaction payload.", "error", err)
			return
		}

		s.logger.Info("[interaction] Handled message_action interaction.",
			slog.Any("payload", payload),
		)

	// https://docs.slack.dev/reference/interaction-payloads/view-interactions-payload/#view_submission
	case "view_submission":
		var payload schema.InteractionPayloadViewSubmission
		if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
			s.logger.Error("[interaction] Error parsing view_submission interaction payload.", "error", err)
			return
		}

		s.logger.Info("[interaction] Handled view_submission interaction.",
			slog.Any("payload", payload),
		)

	// https://docs.slack.dev/reference/interaction-payloads/view-interactions-payload/#view_closed
	case "view_closed":
		var payload schema.InteractionPayloadViewClosed
		if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
			s.logger.Error("[interaction] Error parsing view_closed interaction payload.", "error", err)
			return
		}

		s.logger.Info("[interaction] Handled view_closed interaction.",
			slog.Any("payload", payload),
		)

	default:
		s.logger.Error("[interaction] Unsupported interaction type: " + base.Type)
		return
	}
}

func (s *Service) options(w http.ResponseWriter, r *http.Request) {
	// Assuming the body has been consumed and saved by the prepare middleware.
	body, _ := bodyFromContext(r.Context())

	values, err := url.ParseQuery(string(body))
	if err != nil {
		s.logger.Error("[interaction] Error parsing interaction request.", "error", err)
		http.Error(w, "Failed to parse interaction request.", http.StatusBadRequest)
		return
	}

	rawPayload := values.Get("payload")
	if rawPayload == "" {
		s.logger.Error("[interaction] Missing interaction payload.", "error", err)
		http.Error(w, "Missing interaction payload.", http.StatusBadRequest)
		return
	}

	var base schema.InteractionPayloadBase
	if err := json.Unmarshal([]byte(rawPayload), &base); err != nil {
		s.logger.Error("[interaction] Error parsing interaction payload.", "error", err)
		http.Error(w, "Failed to parse interaction payload.", http.StatusBadRequest)
		return
	}

	if base.Type == "block_suggestion" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(schema.BlockSuggestionResponse{
			OptionGroups: locationOptionGroups,
		})

		s.logger.Info("[interaction] Handled block_suggestion interaction.")
	}
}
