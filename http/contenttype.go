package http

import "fmt"

type CharsetType string

const UTF8 CharsetType = "utf-8"

// ContentType contains predefined mime types.
type ContentType string

const (
	ApplicationJSON  ContentType = "application/json"
	ApplicationXML   ContentType = "application/xml"
	ApplicationXYAML ContentType = "application/x-yaml"
	ApplicationYAML  ContentType = "application/yaml"

	TextJSON  ContentType = "text/json"
	TextXML   ContentType = "text/xml"
	TextXYAML ContentType = "text/x-yaml"
	TextYAML  ContentType = "text/yaml"

	ImageGIF  ContentType = "image/gif"
	ImageJPEG ContentType = "image/jpeg"
	ImagePNG  ContentType = "image/png"
	ImageSVG  ContentType = "image/svg+xml"
	ImageWebP ContentType = "image/webp"
)

func (c ContentType) WithCharset(t CharsetType) ContentTypeWithCharset {
	return ContentTypeWithCharset{
		ContentType: c,
		CharsetType: t,
	}
}

func (c ContentType) Apply(opts *ResponseOptions) {
	opts.ContentType = ContentType(c)
}

type ContentTypeWithCharset struct {
	ContentType ContentType
	CharsetType CharsetType
}

func (c ContentTypeWithCharset) Apply(opts *ResponseOptions) {
	opts.ContentType = ContentType(c.ContentType)
	opts.CharsetType = CharsetType(c.CharsetType)
}

func (c ContentTypeWithCharset) String() string {
	if c.CharsetType == "" {
		return string(c.ContentType)
	}

	return fmt.Sprintf("%s; charset=%s", c.ContentType, c.CharsetType)
}
