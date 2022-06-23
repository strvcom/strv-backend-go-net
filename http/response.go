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
		opts = []ResponseOptions{DefaultResponseOptions}
	}
	o := opts[0]

	w.Header().Set(
		Header.ContentType,
		o.ContentType.WithCharset(o.CharsetType).String(),
	)
	w.WriteHeader(code)

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
	statusCode int,
) {
	w.Header().Set(Header.ContentType, ApplicationJSON.WithCharset(UTF8).String())
	w.WriteHeader(statusCode)

	if err := EncodeJSON(w, r); err != nil {
		panic(fmt.Errorf("reponse encoding: %w", err))
	}
}

func NewErrorResponse(msg string, opts ...ErrorResponseOptions) ErrorResponse {
	r := ErrorResponse{
		Message: msg,
	}
	if len(opts) >= 2 {
		panic("providing more than one option is forbidden")
	}
	if len(opts) == 0 {
		return r
	}

	r.ErrorResponseOptions = opts[0]
	return r
}

type ErrorResponse struct {
	Message string `json:"error_message"`
	ErrorResponseOptions
}

type ErrorResponseOptions struct {
	ErrCode int `json:"error_code,omitempty"`
}
