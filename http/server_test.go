package http

import (
	"context"
	"errors"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.strv.io/net/internal"
)

type cancellableContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *cancellableContext) Err() error {
	return c.ctx.Err()
}

func (c *cancellableContext) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

func (c *cancellableContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *cancellableContext) Cancel() {
	c.cancel()
}

func (c *cancellableContext) CancelFunc() context.CancelFunc {
	return c.cancel
}

func (c *cancellableContext) Value(key any) any {
	return c.ctx.Value(key)
}

func newCancellableContext(ctx context.Context) *cancellableContext {
	cctx, cancel := context.WithCancel(ctx)
	return &cancellableContext{
		ctx:    cctx,
		cancel: cancel,
	}
}

func TestNewServer(t *testing.T) {
	type args struct {
		config *ServerConfig
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "success:default-server-config",
			args: args{
				config: &ServerConfig{},
			},
			want: func() *Server {
				//nolint:gosec
				s := &Server{
					logger:           internal.NewNopLogger(),
					server:           &http.Server{},
					shutdownTimeout:  &defaultShutdownTimeout,
					doBeforeShutdown: []ServerHookFunc(nil),
				}
				s.server.RegisterOnShutdown(s.beforeShutdown)
				return s
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(tt.args.config)
			assert.IsTypef(t, tt.want, s, "NewServer(%v) = %v, want %v", tt.args.config, s, tt.want)
		})
	}
}

func TestServer_Start(t *testing.T) {
	// overrides the default server shutdown timeout for testing purposes.
	defaultShutdownTimeout = 1 * time.Second

	type fields struct {
		server           *http.Server
		signalsListener  chan os.Signal
		shutdownTimeout  *time.Duration
		waitForShutdown  chan struct{}
		doBeforeShutdown []ServerHookFunc
	}
	type args struct {
		ctx *cancellableContext
	}
	tests := []struct {
		name    string
		fields  *fields
		args    args
		testFn  func(*testing.T, args, *fields)
		wantErr error
	}{
		{
			name: "success:start-server-and-cancel-context",
			args: args{ctx: newCancellableContext(context.TODO())},
			testFn: func(t *testing.T, args args, _ *fields) {
				t.Helper()
				args.ctx.Cancel()
			},
			fields: &fields{
				//nolint:gosec
				server:           &http.Server{},
				signalsListener:  make(chan os.Signal, 1),
				waitForShutdown:  make(chan struct{}, 1),
				doBeforeShutdown: []ServerHookFunc{},
				shutdownTimeout:  &defaultShutdownTimeout,
			},
			wantErr: nil,
		},
		{
			name: "success:start-server-and-kill",
			args: args{ctx: newCancellableContext(context.TODO())},
			testFn: func(t *testing.T, _ args, fields *fields) {
				t.Helper()
				fields.signalsListener <- syscall.SIGKILL
			},
			fields: &fields{
				//nolint:gosec
				server:           &http.Server{},
				signalsListener:  make(chan os.Signal, 1),
				waitForShutdown:  make(chan struct{}, 1),
				doBeforeShutdown: []ServerHookFunc{},
				shutdownTimeout:  &defaultShutdownTimeout,
			},
			wantErr: nil,
		},
		{
			name: "success:start-server-and-wait-for-shutdown",
			args: args{ctx: newCancellableContext(context.TODO())},
			testFn: func(t *testing.T, _ args, fields *fields) {
				t.Helper()
				fields.signalsListener <- syscall.SIGKILL
			},
			fields: &fields{
				//nolint:gosec
				server:          &http.Server{},
				signalsListener: make(chan os.Signal, 1),
				waitForShutdown: make(chan struct{}, 1),
				shutdownTimeout: &defaultShutdownTimeout,
				doBeforeShutdown: []ServerHookFunc{
					func(_ context.Context) {
						<-time.After(time.Millisecond * 200)
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nFnCalled := 0
			if len(tt.fields.doBeforeShutdown) > 0 {
				for i, fn := range tt.fields.doBeforeShutdown {
					tt.fields.doBeforeShutdown[i] = func(ctx context.Context) {
						nFnCalled++
						fn(ctx)
					}
				}
			}

			s := &Server{
				logger:           internal.NewNopLogger(),
				server:           tt.fields.server,
				signalsListener:  tt.fields.signalsListener,
				shutdownTimeout:  tt.fields.shutdownTimeout,
				waitForShutdown:  tt.fields.waitForShutdown,
				doBeforeShutdown: tt.fields.doBeforeShutdown,
			}
			s.server.RegisterOnShutdown(s.beforeShutdown)

			errCh := make(chan error, 1)
			go func() {
				deadline, ok := t.Deadline()
				if !ok {
					errCh <- s.Run(tt.args.ctx)
				} else {
					ctxWithDeadline, cancel := context.WithDeadline(tt.args.ctx, deadline)
					defer cancel()
					errCh <- s.Run(ctxWithDeadline)
				}
			}()
			tt.testFn(t, tt.args, tt.fields)

			if err := <-errCh; !errors.Is(err, tt.wantErr) {
				t.Errorf("Server.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, len(tt.fields.doBeforeShutdown), nFnCalled)
		})
	}
}
