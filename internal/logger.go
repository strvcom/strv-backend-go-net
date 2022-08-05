package internal

import "go.strv.io/net/logger"

// NopLogger is a no-op logger that is used if no logger is present.
type NopLogger struct{}

func (l *NopLogger) With(...logger.Field) logger.ServerLogger {
	return l
}

func (*NopLogger) Info(string) {}

func (*NopLogger) Debug(string) {}

func (*NopLogger) Warn(string) {}

func (*NopLogger) Error(string, error) {}
