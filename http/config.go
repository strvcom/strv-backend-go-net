package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"
)

// ServerConfig represents Server configuration.
type ServerConfig struct {
	// Addr is address where HTTP server is listening.
	Addr string `json:"addr"`

	// Handler handles HTTP requests.
	Handler http.Handler `json:"-"`

	// Hooks are server hooks.
	Hooks ServerHooks `json:"-"`

	// Limits are server limits, like timeouts and header restrictions.
	Limits *Limits `json:"limits,omitempty"`
}

// Limits define timeouts and header restrictions.
type Limits struct {
	// Timeouts is a configuration of specific timeouts.
	Timeouts *Timeouts `json:"timeouts,omitempty"`

	// MaxHeaderBytes is part of http.Server.
	// See http.Server for more details.
	MaxHeaderBytes int `json:"maxHeaderBytes"`
}

// Timeouts represents configuration for HTTP server timeouts.
type Timeouts struct {
	// ShutdownTimeout is a timeout before server shutdown.
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// IdleTimeout is part of http.Server.
	// See http.Server for more details.
	IdleTimeout time.Duration `json:"idle_timeout"`

	// ReadTimeout is part of http.Server.
	// See http.Server for more details.
	ReadTimeout time.Duration `json:"read_timeout"`

	// WriteTimeout is part of http.Server.
	// See http.Server for more details.
	WriteTimeout time.Duration `json:"write_timeout"`

	// ReadHeaderTimeout is part of http.Server.
	// See http.Server for more details.
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
}

// TODO: This function might be resued across the codebase, move it to a shared package.
func (t *Timeouts) UnmarshalJSON(data []byte) error {
	var unmarshalledJSON map[string]interface{}

	err := json.Unmarshal(data, &unmarshalledJSON)
	if err != nil {
		return err
	}

	fields := map[string]reflect.Value{}
	rfType := reflect.TypeOf(*t)
	for i := 0; i < rfType.NumField(); i++ {
		jsonTag := rfType.Field(i).Tag.Get("json")
		// Map json tag to field name.
		fields[jsonTag] = reflect.ValueOf(t).Elem().Field(i)
	}

	for k, v := range unmarshalledJSON {
		var duration time.Duration

		switch t := v.(type) {
		case int, int32, int64, float64:
			// By default, interpret number as seconds.
			duration = time.Duration(t.(float64)) * time.Second
		case string:
			duration, err = time.ParseDuration(t)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid type for duration: %#v", unmarshalledJSON)
		}

		f, ok := fields[k]
		if !ok {
			// Shouldn't happen, but just in case.
			return fmt.Errorf("unknown field %q", k)
		}
		f.Set(reflect.ValueOf(duration))
	}

	return nil
}
