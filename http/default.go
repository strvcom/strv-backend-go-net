package http

import (
	"reflect"
	"time"
)

var (
	defaultShutdownTimeout = 30 * time.Second
)

func defaultResponseOptions() ResponseOptions {
	return ResponseOptions{
		EncodeFunc:  EncodeJSON,
		ContentType: ApplicationJSON,
		CharsetType: UTF8,
	}
}

func defaultErrorOptions() ErrorResponseOptions {
	return ErrorResponseOptions{
		ErrCode: "ERR_UNKNOWN",
		ResponseOptions: ResponseOptions{
			EncodeFunc:  EncodeJSON,
			ContentType: ApplicationJSON,
			CharsetType: UTF8,
		},
	}
}

func defaultTo[T any](value T, defaultValue T) T {
	if reflect.ValueOf(value).IsNil() {
		return defaultValue
	}
	return value
}
