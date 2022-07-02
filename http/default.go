package http

import (
	"reflect"
	"time"
)

var (
	defaultResponseOptions = ResponseOptions{
		EncodeFunc:  EncodeJSON,
		ContentType: ApplicationJSON,
		CharsetType: UTF8,
	}
	defaultShutdownTimeout = 30 * time.Second

	defaultErrorCode string = "ERR_UNKNOWN"
)

func defaultTo[T any](value T, defaultValue T) T {
	if reflect.ValueOf(value).IsNil() {
		return defaultValue
	}
	return value
}
