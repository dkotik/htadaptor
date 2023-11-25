package htadaptor

import (
	"errors"
	"net/http"
	"net/textproto"
	"net/url"
)

var ErrNoStringValue = errors.New("empty value")

type StringValueExtractor interface {
	ExtractStringValue(*http.Request) (string, error)
}

type StringValueExtractorFunc func(*http.Request) (string, error)

func (f StringValueExtractorFunc) ExtractStringValue(r *http.Request) (string, error) {
	return f(r)
}

type HeaderValueExtractor string

func NewHeaderValueExtractor(name string) HeaderValueExtractor {
	return HeaderValueExtractor(name)
}

func (e HeaderValueExtractor) ExtractStringValue(r *http.Request) (string, error) {
	value := textproto.MIMEHeader(r.Header).Get(string(e))
	if value == "" {
		return "", ErrNoStringValue
	}
	return value, nil
}

type QueryValueExtractor string

func NewQueryValueExtractor(name string) QueryValueExtractor {
	return QueryValueExtractor(name)
}

func (e QueryValueExtractor) ExtractStringValue(r *http.Request) (value string, err error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return "", err
	}
	value = values.Get(string(e))
	if value == "" {
		return "", ErrNoStringValue
	}
	return value, nil
}

type PathValueExtractor string

func NewPathValueExtractor(name string) PathValueExtractor {
	return PathValueExtractor(name)
}

func (e PathValueExtractor) ExtractStringValue(r *http.Request) (string, error) {
	// value := r.PathValue(string(e))
	// if value == "" {
	//   return "", ErrNoStringValue
	// }
	// return value, nil
	return "", errors.New("requires Go version 1.22")
}
