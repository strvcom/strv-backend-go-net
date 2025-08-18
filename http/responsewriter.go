package http

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
)

type ResponseWriter struct {
	http.ResponseWriter
	statusCode        int
	calledWriteHeader int32
	logger            *slog.Logger
	err               error
	panic             any
}

func NewResponseWriter(w http.ResponseWriter, l *slog.Logger) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter:    w,
		statusCode:        http.StatusOK,
		calledWriteHeader: 0,
		logger:            l,
		err:               nil,
		panic:             nil,
	}
}

func (r *ResponseWriter) StatusCode() int {
	return r.statusCode
}

func (r *ResponseWriter) WriteHeader(statusCode int) {
	if r.TryWriteHeader(statusCode) {
		return
	}
	r.logger.With(
		slog.Int("current_status_code", r.statusCode),
		slog.Int("ignored_status_code", statusCode),
	).WarnContext(context.TODO(), "WriteHeader multiple call")
}

func (r *ResponseWriter) ErrorObject() error {
	return r.err
}

func (r *ResponseWriter) SetErrorObject(err error) {
	r.err = err
}

func (r *ResponseWriter) PanicObject() any {
	return r.panic
}

func (r *ResponseWriter) SetPanicObject(p any) {
	r.panic = p
}

func (r *ResponseWriter) TryWriteHeader(statusCode int) bool {
	if atomic.CompareAndSwapInt32(&r.calledWriteHeader, 0, 1) {
		r.ResponseWriter.WriteHeader(statusCode)
		r.statusCode = statusCode
		return true
	}
	return false
}

func (r *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}
