package http

var (
	defaultResponseOptions = ResponseOptions{
		EncodeFunc:  EncodeJSON,
		ContentType: ApplicationJSON,
		CharsetType: UTF8,
	}

	defaultErrorCode ErrorCode = "ERR_UNKNOWN"
)
