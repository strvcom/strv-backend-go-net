package internal

import (
	"context"
	"log/slog"
)

func NewNopLogger() *slog.Logger {
	return slog.New(nopHandler{})
}

// NopLogger is a no-op logger that discards all of the log messages.
// This logger is used in the case no other logger is provided to the server.
type NopLogger struct {
	slog.Logger
}

type nopHandler struct{}

func (nopHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (nopHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (n nopHandler) WithAttrs([]slog.Attr) slog.Handler {
	return n
}

func (n nopHandler) WithGroup(string) slog.Handler {
	return n
}
