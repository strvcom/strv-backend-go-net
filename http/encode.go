package http

import (
	"encoding/json"
	"net/http"
)

type EncodeFunc func(http.ResponseWriter, any) error

func EncodeJSON(w http.ResponseWriter, data any) error {
	return json.NewEncoder(w).Encode(data)
}

func (EncodeFunc) Apply(o *ResponseOptions) {
	o.EncodeFunc = EncodeJSON
}
