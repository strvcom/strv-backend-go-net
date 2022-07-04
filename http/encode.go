package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// EncodeFunc is a function that encodes data to the response writer.
type EncodeFunc func(w http.ResponseWriter, data any) error

func EncodeJSON(w http.ResponseWriter, data any) error {
	return json.NewEncoder(w).Encode(data)
}

func WithEncodeFunc(fn EncodeFunc) ResponseOption {
	return func(o *ResponseOptions) {
		o.EncodeFunc = fn
	}
}

// DecodeJSON decodes data using JSON marshalling into the type of parameter v.
func DecodeJSON(data any, v any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.NewDecoder(bytes.NewReader(b)).Decode(v)
}

// MustDecodeJSON calls DecodeJSON and panics on error.
func MustDecodeJSON(data any, v any) {
	if err := DecodeJSON(data, v); err != nil {
		panic(fmt.Errorf("decoding: %w", err))
	}
}
