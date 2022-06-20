package http

var (
	// Header contains predefined headers.
	Header = struct {
		AcceptLanguage  string
		ContentLanguage string
		ContentType     string
		WWWAuthenticate string
		XRequestID      string
	}{
		AcceptLanguage:  "Accept-Language",
		ContentLanguage: "Content-Language",
		ContentType:     "Content-Type",
		WWWAuthenticate: "Www-Authenticate",
		XRequestID:      "X-Request-Id",
	}
)
