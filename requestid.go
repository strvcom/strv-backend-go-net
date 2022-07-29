package net

import (
	"context"

	"github.com/google/uuid"
)

type ctxKeyRequestID struct{}

var (
	contextKey = struct {
		requestID ctxKeyRequestID
	}{}
)

func NewRequestID() string {
	return uuid.New().String()
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKey.requestID, requestID)
}

func RequestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(contextKey.requestID).(string)
	if !ok {
		return ""
	}
	return requestID
}
