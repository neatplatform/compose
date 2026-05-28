package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/neatplatform/compose/observability/slack/client"
	"github.com/neatplatform/compose/observability/slack/service"
)

const helpText = `
Set the following environment variables before starting the service:

  NGROK_AUTHTOKEN             ngrok auth token used to expose the service publicly. Required only when the --ngrok flag is set.
  SLACK_APP_SIGNING_SECRET    Slack app signing secret used to verify incoming requests.
  SLACK_APP_AUTH_TOKEN        Slack app OAuth token used to send messages in response to app callbacks.

Endpoints:

  /commands        Handles slash commands
  /events          Handles event subscriptions
  /interactions    Handles interactions
	/options				 Handles options requests

Flags:

`

func main() {
	fs := flag.NewFlagSet("slack", flag.ExitOnError)
	port := fs.Int("port", 8080, "Port to listen on")
	ngrok := fs.Bool("ngrok", false, "Expose the service to the public internet using ngrok (requires NGROK_AUTHTOKEN environment variable)")

	fs.Usage = func() {
		fmt.Fprint(fs.Output(), helpText)
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output())
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
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

	signingSecret := os.Getenv("SLACK_APP_SIGNING_SECRET")
	if signingSecret == "" {
		fmt.Println("SLACK_APP_SIGNING_SECRET environment variable not set")
		os.Exit(1)
	}

	authToken := os.Getenv("SLACK_APP_AUTH_TOKEN")
	if authToken == "" {
		fmt.Println("SLACK_APP_AUTH_TOKEN environment variable not set")
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	// Send a test message with interactive elements, so the user can trigger interactions.
	c := client.New(logger, authToken)
	if err := c.PostMessageFixture(); err != nil {
		logger.Error("Failed to post message fixture.", "error", err)
	}

	// Start the server to handle incoming requests from Slack.
	s := service.New(logger, signingSecret, c)
	s.Start(*port, *ngrok, ngrokAuthToken)
}
