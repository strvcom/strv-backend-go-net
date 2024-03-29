package http

var (
	// Header contains predefined headers.
	Header = struct {
		AcceptLanguage  string
		Authorization   string
		ContentLanguage string
		ContentType     string
		WWWAuthenticate string
		XRequestID      string
		AmazonTraceID   string
	}{
		AcceptLanguage:  "Accept-Language",
		Authorization:   "Authorization",
		ContentLanguage: "Content-Language",
		ContentType:     "Content-Type",
		WWWAuthenticate: "WWW-Authenticate",
		XRequestID:      "X-Request-Id",
		AmazonTraceID:   "X-Amzn-Trace-Id",
	}
)
