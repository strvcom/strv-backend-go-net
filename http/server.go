package http

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	neterrors "go.strv.io/net/errors"
	"go.strv.io/net/internal"
)

func NewServer(config *ServerConfig) *Server {
	if config.Limits == nil {
		config.Limits = &Limits{}
	}

	if config.Logger == nil {
		config.Logger = internal.NewNopLogger()
	}

	s := &Server{
		logger: config.Logger,
		//nolint:gosec // ReadHeaderTimeout is set below
		server: &http.Server{
			Addr:           config.Addr,
			Handler:        config.Handler,
			MaxHeaderBytes: config.Limits.MaxHeaderBytes,
		},
		signalsListener:  make(chan os.Signal, 1),
		shutdownTimeout:  &defaultShutdownTimeout,
		waitForShutdown:  make(chan struct{}, 1),
		doBeforeShutdown: config.Hooks.BeforeShutdown,
	}
	if to := config.Limits.Timeouts; to != nil {
		s.server.ReadTimeout = to.ReadTimeout.Duration()
		s.server.ReadHeaderTimeout = to.ReadHeaderTimeout.Duration()
		s.server.WriteTimeout = to.WriteTimeout.Duration()
		s.server.IdleTimeout = to.IdleTimeout.Duration()

		if to.ShutdownTimeout != nil {
			d := to.ShutdownTimeout.Duration()
			s.shutdownTimeout = &d
		}
	}

	s.server.RegisterOnShutdown(s.beforeShutdown)
	return s
}

type Server struct {
	logger *slog.Logger
	server *http.Server

	signalsListener chan os.Signal
	shutdownTimeout *time.Duration
	waitForShutdown chan struct{}

	doBeforeShutdown []ServerHookFunc
}

// Run calls ListenAndServe but returns error only if err != http.ErrServerClosed.
// Passed context is used as base context of all http requests and to shutdown server gracefully.
func (s *Server) Run(ctx context.Context) error {
	cCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.server.BaseContext = func(_ net.Listener) context.Context {
		return cCtx
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()
	s.logger.Info("server started")

	signal.Notify(s.signalsListener, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			s.logger.Debug("server stopped: server closed")
		} else {
			s.logger.Error("server stopped: error received", slog.Any("error", err))
		}
	case <-ctx.Done():
		s.logger.Error("server stopped: context closed", slog.Any("error", ctx.Err()))
	case sig := <-s.signalsListener:
		s.logger.With(
			slog.Any("signal", sig),
		).Error("server stopped: signal received", slog.Any("error", neterrors.ErrServerInterrupted))
	}

	s.logger.With(
		slog.Duration("timeout", *s.shutdownTimeout),
	).Debug("waiting for server shutdown...")

	if err := s.server.Shutdown(context.Background()); err != nil {
		s.logger.Error("server shutdown", slog.Any("error", err))
		return err
	}
	defer s.logger.Debug("server shutdown complete")

	select {
	case <-s.waitForShutdown:
		return nil
	case <-time.After(*defaultTo(s.shutdownTimeout, &defaultShutdownTimeout)):
		return neterrors.ErrShutdownTimeout
	}
}

func (s *Server) beforeShutdown() {
	if len(s.doBeforeShutdown) == 0 || (s.shutdownTimeout != nil && *s.shutdownTimeout <= 0) {
		s.waitForShutdown <- struct{}{}
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(s.doBeforeShutdown))

	ctx, cancel := context.WithTimeout(context.Background(), *defaultTo(s.shutdownTimeout, &defaultShutdownTimeout))
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
