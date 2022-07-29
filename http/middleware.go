package http

import (
	"net/http"
	"time"

	"go.strv.io/net"
	"go.strv.io/net/internal"
	"go.strv.io/net/logger"
)

func RecoverMiddleware(l logger.ServerLogger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if re := recover(); re != nil {
					l.With(logger.Any("err", re)).Error("panic recover", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func LoggingMiddleware(l logger.ServerLogger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := net.NewRequestID()
			requestStart := time.Now()

			rw := internal.NewResponseWriter(w, l)
			r = r.WithContext(net.WithRequestID(r.Context(), requestID))
			next.ServeHTTP(rw, r)

			ld := LogData{
				Path:               r.URL.EscapedPath(),
				Method:             r.Method,
				RequestID:          requestID,
				Duration:           time.Since(requestStart),
				ResponseStatusCode: rw.StatusCode(),
			}

			WithData(l, ld).Info("request processed")
		})
	}
}

// LogData contains processed request data for purposes of logging.
// Path is path from URL of the request.
// Method is http request method.
// Duration is how long it took to process whole request.
// ResponseStatusCode is http status code which was returned.
// RequestID is unique identifier of request.
type LogData struct {
	Path               string
	Method             string
	Duration           time.Duration
	ResponseStatusCode int
	RequestID          string
}

// WithData returns logger with filled fields based on provided logging settings.
func WithData(l logger.ServerLogger, ld LogData) logger.ServerLogger {
	return l.With(
		logger.Any("method", ld.Method),
		logger.Any("path", ld.Path),
		logger.Any("status_code", ld.ResponseStatusCode),
		logger.Any("request_id", ld.RequestID),
		logger.Any("duration_ms", ld.Duration.Milliseconds()),
	)
}
