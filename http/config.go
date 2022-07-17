package http

import (
	"net/http"

	"go.strv.io/net/logger"

	"go.strv.io/time"
)

// ServerConfig represents Server configuration.
type ServerConfig struct {
	// Addr is address where HTTP server is listening.
	Addr string `json:"addr"`

	// Handler handles HTTP requests.
	Handler http.Handler `json:"-"`

	// Hooks are server hooks.
	Hooks ServerHooks `json:"-"`

	// Limits are server limits, like timeouts and header restrictions.
	Limits *Limits `json:"limits,omitempty"`

	// Logger is server logger.
	Logger logger.ServerLogger
}

// Limits define timeouts and header restrictions.
type Limits struct {
	// Timeouts is a configuration of specific timeouts.
	Timeouts *Timeouts `json:"timeouts,omitempty"`

	// MaxHeaderBytes is part of http.Server.
	// See http.Server for more details.
	MaxHeaderBytes int `json:"maxHeaderBytes"`
}

// Timeouts represents configuration for HTTP server timeouts.
type Timeouts struct {
	// ShutdownTimeout is a timeout before server shutdown.
	//
	// If not provided, the default value is used (30 seconds), if the timeout is less or equal to 0, the server is shutdown immediately,
	// otherwise the server is shutdown after the timeout.
	ShutdownTimeout *time.Duration `json:"shutdown_timeout"`

	// IdleTimeout is part of http.Server.
	// See http.Server for more details.
	IdleTimeout time.Duration `json:"idle_timeout"`

	// ReadTimeout is part of http.Server.
	// See http.Server for more details.
	ReadTimeout time.Duration `json:"read_timeout"`

	// WriteTimeout is part of http.Server.
	// See http.Server for more details.
	WriteTimeout time.Duration `json:"write_timeout"`

	// ReadHeaderTimeout is part of http.Server.
	// See http.Server for more details.
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
}
