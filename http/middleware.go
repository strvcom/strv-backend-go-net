package http

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"go.strv.io/net"
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

type RecoverMiddlewareOptions struct {
	enableStackTrace bool
}

type RecoverMiddlewareOption func(*RecoverMiddlewareOptions)

func WithStackTrace() RecoverMiddlewareOption {
	return func(opts *RecoverMiddlewareOptions) {
		opts.enableStackTrace = true
	}
}

// RecoverMiddleware calls next handler and recovers from a panic.
// If a panic occurs, log this event, set http.StatusInternalServerError as a status code
// and save a panic object into the response writer.
func RecoverMiddleware(l *slog.Logger, opts ...RecoverMiddlewareOption) func(http.Handler) http.Handler {
	options := RecoverMiddlewareOptions{}
	for _, o := range opts {
		o(&options)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if re := recover(); re != nil {
					rw, ok := w.(*ResponseWriter)
					if !ok {
						rw = NewResponseWriter(w, l)
					}

					rw.SetPanicObject(re)
					rw.WriteHeader(http.StatusInternalServerError)

					logAttributes := []slog.Attr{
						slog.String(requestIDLogFieldName, net.RequestIDFromCtx(r.Context())),
						slog.Any("error", re),
					}
					if options.enableStackTrace {
						logAttributes = append(logAttributes, slog.String("stack_trace", string(debug.Stack())))
					}
					l.LogAttrs(r.Context(), slog.LevelError, "panic recover", logAttributes...)
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
//   - Error object if exists
//   - Panic object if exists
//
// If the status code >= http.StatusInternalServerError, logs with error level, info otherwise.
func LoggingMiddleware(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw, ok := w.(*ResponseWriter)
			if !ok {
				rw = NewResponseWriter(w, l)
			}

			requestStart := time.Now()
			next.ServeHTTP(rw, r)
			statusCode := rw.StatusCode()
			requestID := net.RequestIDFromCtx(r.Context())

			ld := RequestData{
				Path:               r.URL.EscapedPath(),
				Method:             r.Method,
				RequestID:          requestID,
				Duration:           time.Since(requestStart),
				ResponseStatusCode: statusCode,
			}

			if statusCode >= http.StatusInternalServerError {
				withRequestData(l, rw, ld).Error("request processed")
			} else {
				withRequestData(l, rw, ld).Info("request processed")
			}
		})
	}
}

// RequestData contains processed request data for logging purposes.
// Path is path from URL of the request.
// Method is HTTP request method.
// Duration is how long it took to process whole request.
// ResponseStatusCode is HTTP status code which was returned.
// RequestID is unique identifier of request.
// Err is error object containing error message.
// Panic is panic object containing error message.
type RequestData struct {
	Path               string
	Method             string
	Duration           time.Duration
	ResponseStatusCode int
	RequestID          string
}

func (r RequestData) LogValue() slog.Value {
	attr := []slog.Attr{
		slog.String("id", r.RequestID),
		slog.String("method", r.Method),
		slog.String("path", r.Path),
		slog.Int("status_code", r.ResponseStatusCode),
		slog.Duration("duration_ms", r.Duration),
	}
	return slog.GroupValue(attr...)
}

// withRequestData returns slog with filled fields.
func withRequestData(l *slog.Logger, rw *ResponseWriter, rd RequestData) *slog.Logger {
	errorObject := rw.ErrorObject()
	panicObject := rw.PanicObject()
	if errorObject != nil {
		l = l.With("error", errorObject)
	}
	if panicObject != nil {
		l = l.With("panic", panicObject)
	}
	return l.With("request", rd)
}
