package htadaptor

import (
	"net/http"
	"net/url"
)

// RequestValueExtractor pulls [url.Values] from an [http.Request]
// in order to provide a [Decoder] with values to populate
// domain request struct with.
type RequestValueExtractor interface {
	ExtractRequestValue(*http.Request) (url.Values, error)
}

// RequestValueExtractorFunc provides a wrapper for functional
// implementation of a [RequestValueExtractor].
type RequestValueExtractorFunc func(*http.Request) (url.Values, error)

// ExtractRequestValue satisfies [RequestValueExtractor] interface
// for [RequestValueExtractorFunc].
func (f RequestValueExtractorFunc) ExtractRequestValue(r *http.Request) (url.Values, error) {
	return f(r)
}
