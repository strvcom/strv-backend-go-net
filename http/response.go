package http

import (
	"fmt"
	"net/http"
)

type ResponseOptions struct {
	EncodeFunc  EncodeFunc
	ContentType ContentType
	CharsetType CharsetType
}

type ResponseOption func(*ResponseOptions)

func WriteResponse(
	w http.ResponseWriter,
	data any,
	code int,
	opts ...ResponseOption,
) error {
	o := defaultResponseOptions
	for _, opt := range opts {
		opt(&o)
	}

	w.Header().Set(
		Header.ContentType,
		o.ContentType.WithCharset(o.CharsetType).String(),
	)
	w.WriteHeader(code)

	if o.EncodeFunc == nil || data == http.NoBody || code == http.StatusNoContent {
		return nil
	}

	if err := o.EncodeFunc(w, data); err != nil {
		return fmt.Errorf("response encoding: %w", err)
	}

	return nil
}

func WithContentType(c ContentType) ResponseOption {
	return func(opts *ResponseOptions) {
		opts.ContentType = c
	}
}

func WithCharsetType(c CharsetType) ResponseOption {
	return func(opts *ResponseOptions) {
		opts.CharsetType = c
	}
}

func WriteErrorResponse(
	w http.ResponseWriter,
	code int,
	opts ...ErrorResponseOption,
) error {
	o := &ErrorResponseOptions{
		ErrCode:         defaultErrorCode,
		ResponseOptions: defaultResponseOptions,
	}
	for _, opt := range opts {
		opt(o)
	}

	w.Header().Set(
		Header.ContentType,
		o.ContentType.WithCharset(o.CharsetType).String(),
	)
	w.WriteHeader(code)

	if o.EncodeFunc == nil {
		return nil
	}

	if err := o.EncodeFunc(w, o); err != nil {
		return fmt.Errorf("response encoding: %w", err)
	}

	return nil
}

type ErrorResponseOptions struct {
	ResponseOptions

	ErrCode string `json:"error_code"`
	ErrData any    `json:"error_data,omitempty"`
}

type ErrorResponseOption func(*ErrorResponseOptions)

func WithErrorCode(code string) ErrorResponseOption {
	return func(o *ErrorResponseOptions) {
		o.ErrCode = code
	}
}

func WithErrorData(data any) ErrorResponseOption {
	return func(o *ErrorResponseOptions) {
		o.ErrData = data
	}
}
