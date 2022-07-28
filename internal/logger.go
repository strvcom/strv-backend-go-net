package internal

import "go.strv.io/net/logger"

// NopLogger is a no-op logger that is used if no logger is present.
type NopLogger struct{}

func (l *NopLogger) With(...logger.Field) logger.ServerLogger {
	return l
}

func (*NopLogger) Info(string) {
	return
}

func (*NopLogger) Debug(string) {
	return
}

func (*NopLogger) Warn(string) {
	return
}

func (*NopLogger) Error(string, error) {
	return
}
