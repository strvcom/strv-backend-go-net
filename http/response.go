package http

import (
	"fmt"
	"net/http"

	"go.strv.io/net/internal"
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
	statusCode int,
	opts ...ResponseOption,
) error {
	o := defaultResponseOptions()
	for _, opt := range opts {
		opt(&o)
	}

	w.Header().Set(
		Header.ContentType,
		o.ContentType.WithCharset(o.CharsetType).String(),
	)
	w.WriteHeader(statusCode)

	if o.EncodeFunc == nil || data == http.NoBody || statusCode == http.StatusNoContent {
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
	statusCode int,
	opts ...ErrorResponseOption,
) error {
	o := defaultErrorOptions()
	for _, opt := range opts {
		opt(&o)
	}

	w.Header().Set(
		Header.ContentType,
		o.ContentType.WithCharset(o.CharsetType).String(),
	)
	w.WriteHeader(statusCode)

	if o.EncodeFunc == nil {
		return nil
	}

	if err := o.EncodeFunc(w, o); err != nil {
		return fmt.Errorf("response encoding: %w", err)
	}

	if rw, ok := w.(*internal.ResponseWriter); ok {
		rw.SetErrorObject(o.Err)
	}

	return nil
}

type ErrorResponseOptions struct {
	ResponseOptions `json:"-"`

	RequestID string `json:"requestId,omitempty"`

	Err        error  `json:"-"`
	ErrCode    string `json:"errorCode"`
	ErrMessage string `json:"errorMessage,omitempty"`
	ErrData    any    `json:"errorData,omitempty"`
}

type ErrorResponseOption func(*ErrorResponseOptions)

func WithRequestID(id string) ErrorResponseOption {
	return func(o *ErrorResponseOptions) {
		o.RequestID = id
	}
}

func WithError(err error) ErrorResponseOption {
	return func(o *ErrorResponseOptions) {
		o.Err = err
	}
}

func WithErrorCode(code string) ErrorResponseOption {
	return func(o *ErrorResponseOptions) {
		o.ErrCode = code
	}
}

func WithErrorMessage(msg string) ErrorResponseOption {
	return func(o *ErrorResponseOptions) {
		o.ErrMessage = msg
	}
}

func WithErrorData(data any) ErrorResponseOption {
	return func(o *ErrorResponseOptions) {
		o.ErrData = data
	}
}
