package http

import (
	"fmt"
	"net/http"
)

var DefaultResponseOptions = &ResponseOptions{
	EncodeFunc:  EncodeJSON,
	ContentType: ApplicationJSON,
	CharsetType: UTF8,
}

type ResponseOptions struct {
	EncodeFunc  EncodeFunc
	ContentType ContentType
	CharsetType CharsetType
}

func WriteResponse(
	w http.ResponseWriter,
	data any,
	code int,
	opts ...ResponseOptions,
) {
	if len(opts) >= 2 {
		panic("providing more than one option is forbidden")
	}
	if len(opts) == 0 {
		opts = []ResponseOptions{*DefaultResponseOptions}
	}
	o := opts[0]

	w.Header().Set(Header.ContentType, o.ContentType.WithCharset(UTF8).String())
	w.WriteHeader(code)

	if o.EncodeFunc == nil || data == http.NoBody || code == http.StatusNoContent {
		return
	}

	err := o.EncodeFunc(w, data)
	if err != nil {
		panic(fmt.Errorf("unable to encoded response: %w", err))
	}
}

func WriteErrorResponse(
	w http.ResponseWriter,
	r ErrorResponse,
	code int,
) {
	w.Header().Set(Header.ContentType, ApplicationJSON.WithCharset(UTF8).String())
	w.WriteHeader(code)

	err := EncodeJSON(w, r)
	if err != nil {
		panic(fmt.Errorf("unable to encode response: %w", err))
	}
}

func NewErrorResponse(msg string, opts ...ErrorResponseOptions) ErrorResponse {
	if len(opts) >= 2 {
		panic("providing more than one option is forbidden")
	}
	if len(opts) == 0 {
		opts = []ErrorResponseOptions{}
	}
	return ErrorResponse{
		Message:              msg,
		ErrorResponseOptions: opts[0],
	}
}

type ErrorResponse struct {
	Message string `json:"error_message"`
	ErrorResponseOptions
}

type ErrorResponseOptions struct {
	ErrCode int `json:"error_code,omitempty"`
}
