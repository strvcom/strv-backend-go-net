package internal

import (
	"net/http"
	"sync/atomic"

	"go.strv.io/net/logger"
)

type ResponseWriter struct {
	http.ResponseWriter
	statusCode        int
	calledWriteHeader int32
	logger            logger.ServerLogger
	panic             any
}

func NewResponseWriter(w http.ResponseWriter, l logger.ServerLogger) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter:    w,
		statusCode:        http.StatusOK,
		calledWriteHeader: 0,
		logger:            l,
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
		logger.Any("current_status_code", r.statusCode),
		logger.Any("ignored_status_code", statusCode),
	).Warn("WriteHeader multiple call")
}

func (r *ResponseWriter) PanicObject() any {
	return r.panic
}

func (r *ResponseWriter) SetPanicObject(panic any) {
	r.panic = panic
}

func (r *ResponseWriter) TryWriteHeader(statusCode int) bool {
	if atomic.CompareAndSwapInt32(&r.calledWriteHeader, 0, 1) {
		r.ResponseWriter.WriteHeader(statusCode)
		r.statusCode = statusCode
		return true
	}
	return false
}
