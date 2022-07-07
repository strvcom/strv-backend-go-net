package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteResponse(t *testing.T) {
	type args struct {
		w    *httptest.ResponseRecorder
		data any
		code int
		opts []ResponseOption
	}
	tests := []struct {
		name   string
		args   args
		testFn func(*testing.T, args)
	}{
		{
			name: "success:default-no-content",
			args: args{
				w:    httptest.NewRecorder(),
				data: http.NoBody,
				code: http.StatusNoContent,
			},
			testFn: func(t *testing.T, args args) {
				contentType := defaultResponseOptions().ContentType
				charsetType := defaultResponseOptions().CharsetType

				assert.Equal(t, args.w.Code, http.StatusNoContent)
				assert.Equal(t, args.w.Body, bytes.NewBuffer(nil))
				assert.Equal(
					t,
					args.w.Header().Get(Header.ContentType),
					contentType.WithCharset(charsetType).String(),
				)
			},
		},
		{
			name: "success:no-content-image/gif-utf8",
			args: args{
				w:    httptest.NewRecorder(),
				data: http.NoBody,
				code: http.StatusNoContent,
				opts: []ResponseOption{WithContentType(ImageGIF)},
			},
			testFn: func(t *testing.T, args args) {
				contentType := ImageGIF
				charsetType := defaultResponseOptions().CharsetType

				assert.Equal(t, args.w.Code, http.StatusNoContent)
				assert.Equal(t, args.w.Body, bytes.NewBuffer(nil))
				assert.Equal(
					t,
					args.w.Header().Get(Header.ContentType),
					contentType.WithCharset(charsetType).String(),
				)
			},
		},
		{
			name: "success:no-content-image/gif-custom-charset",
			args: args{
				w:    httptest.NewRecorder(),
				data: http.NoBody,
				code: http.StatusNoContent,
				opts: []ResponseOption{
					WithContentType(ImageGIF),
					WithCharsetType("custom"),
				},
			},
			testFn: func(t *testing.T, args args) {
				contentType := ImageGIF
				charsetType := CharsetType("custom")

				assert.Equal(t, args.w.Code, http.StatusNoContent)
				assert.Equal(t, args.w.Body, bytes.NewBuffer(nil))
				assert.Equal(
					t,
					args.w.Header().Get(Header.ContentType),
					contentType.WithCharset(charsetType).String(),
				)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteResponse(tt.args.w, tt.args.data, tt.args.code, tt.args.opts...)
			tt.testFn(t, tt.args)
		})
	}
}

func TestWriteErrorResponse(t *testing.T) {
	type args struct {
		w    *httptest.ResponseRecorder
		code int
		opts []ErrorResponseOption
	}
	tests := []struct {
		name   string
		args   args
		testFn func(*testing.T, args)
	}{
		{
			name: "success:default-error-code",
			args: args{
				w:    httptest.NewRecorder(),
				code: http.StatusInternalServerError,
			},
			testFn: func(t *testing.T, args args) {
				assert.Equal(t, args.w.Code, http.StatusInternalServerError)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteErrorResponse(tt.args.w, tt.args.code, tt.args.opts...)
			tt.testFn(t, tt.args)
		})
	}
}
