package http

import (
	"encoding/json"
	"net/http"
)

// EncodeFunc is a function that encodes data to the response writer.
type EncodeFunc func(w http.ResponseWriter, data any) error

func EncodeJSON(w http.ResponseWriter, data any) error {
	return json.NewEncoder(w).Encode(data)
}

func (EncodeFunc) Apply(o *ResponseOptions) {
	o.EncodeFunc = EncodeJSON
}
