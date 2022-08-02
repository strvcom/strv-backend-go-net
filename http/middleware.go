package http

import (
	"net/http"
	"time"

	"go.strv.io/net"
	"go.strv.io/net/internal"
	"go.strv.io/net/logger"
)

const (
	requestIDLogFieldName = "request_id"
)

// RequestIDFunc is used for obtaining a request ID from the HTTP header.
type RequestIDFunc func(h http.Header) string

// RequestIDMiddleware saves request ID into the request context.
// If context already contains request ID, next handler is called.
// If the user provided function returns empty request ID, a new one is generated.
func RequestIDMiddleware(f RequestIDFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requestID := net.RequestIDFromCtx(r.Context()); requestID != "" {
				next.ServeHTTP(w, r)
				return
			}

			var requestID string
			if rID := f(r.Header); rID != "" {
				requestID = rID
			} else {
				requestID = net.NewRequestID()
			}

			ctx := net.WithRequestID(r.Context(), requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RecoverMiddleware calls next handler and recovers from a panic.
// If a panic occurs, log this event, set http.StatusInternalServerError as a status code
// and save a panic object into the response writer.
func RecoverMiddleware(l logger.ServerLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if re := recover(); re != nil {
					rw, ok := w.(*internal.ResponseWriter)
					if !ok {
						rw = internal.NewResponseWriter(w, l)
					}

					rw.SetPanicObject(re)
					rw.WriteHeader(http.StatusInternalServerError)

					l.With(
						logger.Any("err", re),
						logger.Any(requestIDLogFieldName, net.RequestIDFromCtx(r.Context())),
					).Error("panic recover", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs:
//   - URL path
//   - HTTP method
//   - Request ID
//   - Duration of a request
//   - HTTP status code
//   - Panic object if exists
// If the status code >= http.StatusInternalServerError, logs with error level, info otherwise.
func LoggingMiddleware(l logger.ServerLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw, ok := w.(*internal.ResponseWriter)
			if !ok {
				rw = internal.NewResponseWriter(w, l)
			}

			requestStart := time.Now()
			next.ServeHTTP(rw, r)
			statusCode := rw.StatusCode()
			requestID := net.RequestIDFromCtx(r.Context())

			ld := LogData{
				Path:               r.URL.EscapedPath(),
				Method:             r.Method,
				RequestID:          requestID,
				Duration:           time.Since(requestStart),
				ResponseStatusCode: statusCode,
				Panic:              rw.PanicObject(),
			}

			if statusCode >= http.StatusInternalServerError {
				WithData(l, ld).Error("request processed", nil)
			} else {
				WithData(l, ld).Info("request processed")
			}
		})
	}
}

// LogData contains processed request data for logging purposes.
// Path is path from URL of the request.
// Method is HTTP request method.
// Duration is how long it took to process whole request.
// ResponseStatusCode is HTTP status code which was returned.
// RequestID is unique identifier of request.
// Panic is panic object containing error message.
type LogData struct {
	Path               string
	Method             string
	Duration           time.Duration
	ResponseStatusCode int
	RequestID          string
	Panic              any
}

// WithData returns logger with filled fields.
func WithData(l logger.ServerLogger, ld LogData) logger.ServerLogger {
	l = l.With(
		logger.Any("method", ld.Method),
		logger.Any("path", ld.Path),
		logger.Any("status_code", ld.ResponseStatusCode),
		logger.Any("request_id", ld.RequestID),
		logger.Any("duration_ms", ld.Duration.Milliseconds()),
	)
	if ld.Panic != nil {
		l = l.With(logger.Any("panic", ld.Panic))
	}
	return l
}
