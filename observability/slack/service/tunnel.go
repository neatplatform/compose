package service

import (
	"context"
	"fmt"
	"log/slog"

	"golang.ngrok.com/ngrok/v2"
)

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

	t.logger.Info("ngrok tunnel established",
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
