package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	neterrors "go.strv.io/net/errors"
	"go.strv.io/net/internal"
)

type Server struct {
	logger *slog.Logger
	server *http.Server

	signalsListener chan os.Signal
	shutdownTimeout *time.Duration
	waitForShutdown chan struct{}

	doBeforeShutdown []ServerHookFunc
}

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

// Run calls ListenAndServe but returns error only if err != http.ErrServerClosed.
// Server is shutdown when passed context is canceled, or when SIGTERM is received.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()
	s.logger.InfoContext(ctx, "server started")

	signal.Notify(s.signalsListener, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			s.logger.DebugContext(ctx, "server stopped: server closed")
		} else {
			s.logger.ErrorContext(ctx, "server stopped: error received", slog.Any("error", err))
		}
	case <-ctx.Done():
		s.logger.InfoContext(ctx, "server stopped: context closed", slog.Any("error", ctx.Err()))
	case sig := <-s.signalsListener:
		s.logger.With(
			slog.Any("signal", sig),
		).InfoContext(ctx, "server stopped: signal received", slog.Any("error", neterrors.ErrServerInterrupted))
	}

	s.logger.With(
		slog.Duration("timeout", *s.shutdownTimeout),
	).DebugContext(ctx, "waiting for server shutdown...")

	if err := s.server.Shutdown(context.Background()); err != nil {
		s.logger.ErrorContext(ctx, "server shutdown", slog.Any("error", err))
		return err
	}
	defer s.logger.DebugContext(ctx, "server shutdown complete")

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
	// Each ServerHookFunc will be run in parallel with the main http.Server.Shutdown(). Server.Run() will block
	// until Shutdown() and all BeforeShutdown hooks completes (or ShutdownTimeout passes).
	// Passed context is canceled after ShutdownTimeout passes, but at that point, completion of the hook
	// is not waited for anymore (as Run returns after such timeout).
	BeforeShutdown []ServerHookFunc
}

type ServerHookFunc func(context.Context)
