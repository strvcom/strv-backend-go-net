package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

var defaultLimits = &Limits{}

func NewServer(config *ServerConfig) *Server {
	if config.Limits == nil {
		config.Limits = defaultLimits
	}

	s := &Server{
		server: http.Server{
			Addr:           config.Addr,
			Handler:        config.Handler,
			MaxHeaderBytes: config.Limits.MaxHeaderBytes,
		},
		doBeforeShutdown: config.Hooks.BeforeShutdown,
	}
	if to := config.Limits.Timeouts; to != nil {
		s.server.ReadTimeout = to.ReadTimeout
		s.server.ReadHeaderTimeout = to.ReadHeaderTimeout
		s.server.WriteTimeout = to.WriteTimeout
		s.server.IdleTimeout = to.IdleTimeout

		s.shutdownTimeout = to.ShutdownTimeout
	}

	return s
}

type Server struct {
	server http.Server

	shutdownTimeout time.Duration
	waitForShutdown chan struct{}

	doBeforeShutdown []ServerHookFunc
}

// Start calls ListenAndServe but returns error only if err != http.ErrServerClosed.
// Passed context is used as base context of all http request and to shutdown server gracefully.
func (s *Server) Start(ctx context.Context) error {
	s.server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	go func() {
		<-ctx.Done()
		s.waitForShutdown = make(chan struct{})
		s.beforeShutdown(ctx)
		s.waitForShutdown <- struct{}{}
	}()
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	if s.waitForShutdown != nil {
		if s.shutdownTimeout > 0 {
			select {
			case <-s.waitForShutdown:
				return nil
			case <-time.After(s.shutdownTimeout):
				return fmt.Errorf("shutdown timeout %s exceeded", s.shutdownTimeout)
			}
		}
		<-s.waitForShutdown
	}
	return nil
}

func (s *Server) beforeShutdown(ctx context.Context) {
	wg := &sync.WaitGroup{}
	wg.Add(len(s.doBeforeShutdown))

	for _, f := range s.doBeforeShutdown {
		go func(f ServerHookFunc, wg *sync.WaitGroup) {
			f(ctx)
			wg.Done()
		}(f, wg)
	}
	wg.Wait()
}

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
}

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
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

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

type ServerHooks struct {
	BeforeShutdown []ServerHookFunc
}

type ServerHookFunc func(context.Context) error
