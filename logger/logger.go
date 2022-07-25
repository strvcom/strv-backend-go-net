package logger

// ServerLogger is an interface for logging server messages.
type ServerLogger interface {
	// With creates a child logger and adds structured context to it. Fields added
	// to the child don't affect the parent, and vice versa.
	With(fields ...Field) ServerLogger

	// Info logs a message at InfoLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Info(msg string)

	// Debug logs a message at DebugLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Debug(msg string)

	// Warn logs a message at WarnLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Warn(msg string)

	// Error logs a message at ErrorLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Error(msg string, err error)
}

// Field contains logging context passed to logging functions.
// It is inspired by zapcore.Field. See https://godoc.org/go.uber.org/zap#Field for details.
type Field struct {
	Key   string
	Value any
}

func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}
