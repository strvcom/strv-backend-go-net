package errors

import "errors"

var (
	ErrShutdownTimeout   = errors.New("server shutdown timeout")
	ErrServerInterrupted = errors.New("server interrupted")
)
