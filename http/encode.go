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

func (f EncodeFunc) Apply(o *ResponseOptions) {
	o.EncodeFunc = f
}

func DecodeJSON(data any, v interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.NewDecoder(bytes.NewReader(b)).Decode(v)
}

func MustDecodeJSON(data any) {
	if err := DecodeJSON(data, nil); err != nil {
		panic(fmt.Errorf("decoding: %w", err))
	}
}
