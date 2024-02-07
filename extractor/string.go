package extractor

import "net/http"

type StringValueExtractor interface {
	ExtractStringValue(*http.Request) (string, error)
}

type StringValueExtractorFunc func(*http.Request) (string, error)

func (f StringValueExtractorFunc) ExtractStringValue(r *http.Request) (string, error) {
	return f(r)
}
