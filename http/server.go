package http

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
		signalsListener:  make(chan os.Signal, 1),
		waitForShutdown:  make(chan struct{}, 1),
		doBeforeShutdown: config.Hooks.BeforeShutdown,
	}
	if to := config.Limits.Timeouts; to != nil {
		s.server.ReadTimeout = to.ReadTimeout
		s.server.ReadHeaderTimeout = to.ReadHeaderTimeout
		s.server.WriteTimeout = to.WriteTimeout
		s.server.IdleTimeout = to.IdleTimeout

		s.shutdownTimeout = to.ShutdownTimeout
	}

	s.server.RegisterOnShutdown(s.beforeShutdown)
	return s
}

type Server struct {
	server http.Server

	signalsListener chan os.Signal
	shutdownTimeout time.Duration
	waitForShutdown chan struct{}

	doBeforeShutdown []ServerHookFunc
}

// Start calls ListenAndServe but returns error only if err != http.ErrServerClosed.
// Passed context is used as base context of all http requests and to shutdown server gracefully.
func (s *Server) Start(ctx context.Context) error {
	cCtx, cancel := context.WithCancel(ctx)
	s.server.BaseContext = func(_ net.Listener) context.Context {
		return cCtx
	}

	signal.Notify(s.signalsListener, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case <-errCh:
		// TODO: Log error.
	case <-ctx.Done():
		// TODO: Log context closed.
	case <-s.signalsListener:
		// TODO: Log signal received.
	}
	cancel()

	// TODO: Log server shutdown error.
	_ = s.server.Shutdown(context.Background())
	select {
	case <-s.waitForShutdown:
		return nil
	case <-time.After(s.shutdownTimeout):
		return ErrShutdownTimeout
	}
}

func (s *Server) beforeShutdown() {
	wg := &sync.WaitGroup{}
	wg.Add(len(s.doBeforeShutdown))

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	for _, f := range s.doBeforeShutdown {
		go func(f ServerHookFunc, wg *sync.WaitGroup) {
			f(ctx)
			wg.Done()
		}(f, wg)
	}
	wg.Wait()
	s.waitForShutdown <- struct{}{}
}

type ServerHooks struct {
	BeforeShutdown []ServerHookFunc
}

type ServerHookFunc func(context.Context)
