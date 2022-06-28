package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteResponse(t *testing.T) {
	type args struct {
		w    *httptest.ResponseRecorder
		data any
		code HTTPStatusCode
		opts []ResponseOption
	}
	tests := []struct {
		name     string
		args     args
		testFunc func(*testing.T, args)
	}{
		{
			name: "success:no-content",
			args: args{
				w:    httptest.NewRecorder(),
				data: http.NoBody,
				code: http.StatusNoContent,
			},
			testFunc: func(t *testing.T, args args) {
				assert.Equal(t, args.w.Body, http.NoBody)
				assert.Equal(t, args.w.Code, http.StatusNoContent)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteResponse(tt.args.w, tt.args.data, tt.args.code, tt.args.opts...)
		})
		tt.testFunc(t, tt.args)
	}
}

func TestWriteErrorResponse(t *testing.T) {
	type args struct {
		w    *httptest.ResponseRecorder
		r    ErrorResponse
		code HTTPStatusCode
	}
	tests := []struct {
		name     string
		args     args
		testFunc func(*testing.T, args)
	}{
		{
			name: "success:default-error-code",
			args: args{
				w:    httptest.NewRecorder(),
				r:    ErrorResponse{},
				code: http.StatusOK,
			},
			testFunc: func(t *testing.T, args args) {
				assert.Equal(t, args.w.Code, DefaultErrorCode)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteErrorResponse(tt.args.w, tt.args.r, tt.args.code)
		})
		tt.testFunc(t, tt.args)
	}
}
