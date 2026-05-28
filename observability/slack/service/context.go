package service

import "context"

type contextKey string

const contextKeyBody contextKey = "body"

func contextWithBody(ctx context.Context, body []byte) context.Context {
	return context.WithValue(ctx, contextKeyBody, body)
}

func bodyFromContext(ctx context.Context) ([]byte, bool) {
	body, ok := ctx.Value(contextKeyBody).([]byte)
	return body, ok
}
