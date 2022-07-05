package http

import "errors"

var ErrShutdownTimeout = errors.New("http: server shutdown timeout")
