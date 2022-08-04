# STRV net

Go package facilitating writing API applications in a fast and easy manner.

## Available packages

### errors
Definition of common errors.

### logger
Interface `ServerLogger` implements common logging methods.

### net
Common functionality that comes in handy regardless of the used API architecture. `net` currently supports generating request IDs with some helper methods.

### http
Wrapper around the Go native http server. `http` defines the `Server` that can be configured by the `ServerConfig`. Implemented features:
- Started http server can be easily stopped by cancelling the context that is passed by the `Run` method.
- The `Server` can be configured with a logger for logging important information during starting/ending of the server.
- The `Server` listens for `SIGINT` and `SIGTERM` signals so it can be stopped by firing the signal.
- By the `ServerConfig` can be configured functions to be called before the `Server` ends.

`http` defines several helper consctructs:
- Content types and headers which are frequently used by APIs.
- Middlewares:
	- `RequestIDMiddleware` sets request id in to the context.
	- `RecoverMiddleware` recovers from panic and sets panic object into the response writer for logging.
	- `LoggingMiddleware` logs information about the request (method, path, status code, request id and duration of the request).
- Method `WriteResponse` for writing a http response and `WriteErrorResponse` for writing an error http response. Writing of responses can be configured by `ResponseOption`.

## Examples
### http
Starting the server:
```go
package main

import (
	...

	httpx "go.strv.io/net/http"
)

func main() {
	...
	
	serverConfig := httpx.ServerConfig{
		Addr:    ":8080",
		Handler: handler(), // define your http handler
		Hooks: httpx.ServerHooks{
			BeforeShutdown: []httpx.ServerHookFunc{
				func(_ context.Context) {
					storage.Close() // it may be useful for example to close a storage before the server ends
				},
			},
		},
		Limits: nil,
		Logger: util.NewServerLogger("httpx.Server"), // wrapper around zap logger to implement httpx logging interface
	}
	server := httpx.NewServer(&serverConfig)
	if err = server.Start(ctx); err != nil {
		logger.Fatal("HTTP server unexpectedly ended", zap.Error(err))
	}
}
```

Writing http responses:
```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r.Context())
	
	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		_ = httpx.WriteErrorResponse(w, http.StatusInternalServerError, httpx.WithErrorCode("ERR_UNKNOWN"))
		return
	}
	
	userResp := model.ToUser(user)
	_ = httpx.WriteResponse(w, userResp, http.StatusOK)
}
```
