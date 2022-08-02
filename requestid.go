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

// NewRequestID returns generated UUID.
func NewRequestID() string {
	return uuid.New().String()
}

// WithRequestID saves request ID into the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKey.requestID, requestID)
}

// RequestIDFromCtx extracts request ID from the context.
func RequestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(contextKey.requestID).(string)
	if !ok {
		return ""
	}
	return requestID
}
