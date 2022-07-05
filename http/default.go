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

	defaultErrorOptions = ErrorResponseOptions{
		ErrCode: "ERR_UNKNOWN",
		ResponseOptions: ResponseOptions{
			EncodeFunc:  EncodeJSON,
			ContentType: ApplicationJSON,
			CharsetType: UTF8,
		},
	}
)

func defaultTo[T any](value T, defaultValue T) T {
	if reflect.ValueOf(value).IsNil() {
		return defaultValue
	}
	return value
}
