package http

import (
	"reflect"
	"time"
)

var (
	defaultShutdownTimeout = 30 * time.Second
	defaultErrCode         = "ERR_UNKNOWN"
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
		ResponseOptions: ResponseOptions{
			EncodeFunc:  EncodeJSON,
			ContentType: ApplicationJSON,
			CharsetType: UTF8,
		},
		Err:     nil,
		ErrCode: defaultErrCode,
		ErrData: nil,
	}
}

func defaultTo[T any](value T, defaultValue T) T {
	if reflect.ValueOf(value).IsNil() {
		return defaultValue
	}
	return value
}
