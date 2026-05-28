package http

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func RunWebhook(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	port := fs.Int("port", 8080, "Port to listen on")
	endpoint := fs.String("endpoint", "/webhook", "HTTP path to receive requests on")
	ngrok := fs.Bool("ngrok", false, "Expose the service to the public internet using ngrok (requires NGROK_AUTHTOKEN environment variable)")

	if err := fs.Parse(args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	if *port <= 0 || *port > 65535 {
		fmt.Println("Invalid port number.")
		os.Exit(1)
	}

	ngrokAuthToken := os.Getenv("NGROK_AUTHTOKEN")
	if *ngrok && ngrokAuthToken == "" {
		fmt.Println("NGROK_AUTHTOKEN environment variable not set")
		os.Exit(1)
	}

	newWebhookService(*endpoint).Start(*port, *ngrok, ngrokAuthToken)
}

// webhookService is an HTTP webhook service that receives and logs incoming payloads.
type webhookService struct {
	*service
}

func newWebhookService(endpoint string) *webhookService {
	s := &webhookService{
		service: newService(),
	}

	s.RegisterRoute(endpoint, s.webhook)

	return s
}

func (s *webhookService) webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		s.logger.Error("Unexpected HTTP method: " + r.Method)
		http.Error(w, "Only POST and PUT are allowed.", http.StatusMethodNotAllowed)
		return
	}

	headers := make([]string, 0, len(r.Header))
	for h, vs := range r.Header {
		headers = append(headers, fmt.Sprintf("%s:%s", h, strings.Join(vs, ",")))
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyLen))
	if err != nil {
		s.logger.Error("Error reading request body.", "error", err)
		http.Error(w, "Failed to read request body.", http.StatusInternalServerError)
		return
	}

	// Log the received payload.
	s.logger.Debug("Received payload.",
		slog.String("headers", strings.Join(headers, " ")),
		slog.String("body", string(body)),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "OK")
}
