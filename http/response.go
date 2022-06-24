package http

import (
	"fmt"
	"net/http"
)

var DefaultResponseOptions = ResponseOptions{
	EncodeFunc:  EncodeJSON,
	ContentType: ApplicationJSON,
	CharsetType: UTF8,
}

type ResponseOptions struct {
	EncodeFunc  EncodeFunc
	ContentType ContentType
	CharsetType CharsetType
}

type ResponseOption interface {
	Apply(*ResponseOptions)
}

type HTTPStatusCode int

func WriteResponse(
	w http.ResponseWriter,
	data any,
	code HTTPStatusCode,
	opts ...ResponseOption,
) {
	o := &ResponseOptions{}
	for _, opt := range opts {
		opt.Apply(o)
	}

	w.Header().Set(
		Header.ContentType,
		o.ContentType.WithCharset(o.CharsetType).String(),
	)
	w.WriteHeader(int(code))

	if o.EncodeFunc == nil || data == http.NoBody || code == http.StatusNoContent {
		return
	}

	if err := o.EncodeFunc(w, data); err != nil {
		panic(fmt.Errorf("response encoding: %w", err))
	}
}

func WriteErrorResponse(
	w http.ResponseWriter,
	r ErrorResponse,
	code HTTPStatusCode,
) {
	w.Header().Set(Header.ContentType, ApplicationJSON.WithCharset(UTF8).String())
	w.WriteHeader(int(code))

	if err := EncodeJSON(w, r); err != nil {
		panic(fmt.Errorf("reponse encoding: %w", err))
	}
}

func NewErrorResponse(msg string, opts ...ErrorResponseOption) ErrorResponse {
	r := ErrorResponse{
		Message: msg,
	}
	for _, opt := range opts {
		opt.Apply(&r)
	}

	return r
}

type ErrorCode int

func (c ErrorCode) Apply(r *ErrorResponse) {
	r.ErrCode = c
}

type ErrorResponse struct {
	Message string    `json:"error_message"`
	ErrCode ErrorCode `json:"error_code,omitempty"`
}

type ErrorResponseOption interface {
	Apply(*ErrorResponse)
}
