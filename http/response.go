package http

import (
	"errors"
	"fmt"
	"net/http"
)

type ResponseOptions struct {
	EncodeFunc  EncodeFunc
	ContentType ContentType
	CharsetType CharsetType
}

type ResponseOption interface {
	Apply(*ResponseOptions)
}

func WriteResponse(
	w http.ResponseWriter,
	data any,
	code int,
	opts ...ResponseOption,
) error {
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
		return nil
	}

	if err := o.EncodeFunc(w, data); err != nil {
		return fmt.Errorf("response encoding: %w", err)
	}

	return nil
}

func WriteErrorResponse(
	w http.ResponseWriter,
	r ErrorResponse,
	statusCode int,
) error {
	if r.ErrCode == "" {
		r.ErrCode = defaultErrorCode
	}

	w.Header().Set(Header.ContentType, ApplicationJSON.WithCharset(UTF8).String())
	w.WriteHeader(statusCode)

	if err := EncodeJSON(w, r); err != nil {
		return fmt.Errorf("reponse encoding: %w", err)
	}

	return nil
}

func NewErrorResponse(errCode ErrorCode, opts ...ErrorResponseOption) ErrorResponse {
	r := ErrorResponse{
		ErrCode: errCode,
	}
	for _, opt := range opts {
		opt.Apply(&r)
	}

	return r
}

type ErrorCode string

type ErrorData map[string]any

func (d ErrorData) Apply(r *ErrorResponse) {
	r.ErrData = d
}

func (d *ErrorData) Unmarshal(data any) error {
	if data == nil {
		return errors.New("data must not be empty")
	}

	if e, ok := data.(ErrorData); ok {
		*d = e
	}
	if e, ok := data.(map[string]any); ok {
		*d = ErrorData(e)
	}

	if err := DecodeJSON(data, d); err != nil {
		return fmt.Errorf("decoding: %w", err)
	}

	return nil
}

type ErrorResponse struct {
	ErrCode ErrorCode `json:"error_code"`
	ErrData any       `json:"error_data,omitempty"`
}

type ErrorResponseOption interface {
	Apply(*ErrorResponse)
}
