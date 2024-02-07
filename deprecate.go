package htadaptor

// TODO: better extractors are in reflectd package. All the below functions must be removed in favor of reflectd.

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

type CookieValueExtractor string

func NewCookieValueExtractor(name string) CookieValueExtractor {
	return CookieValueExtractor(name)
}

func (e CookieValueExtractor) ExtractStringValue(r *http.Request) (string, error) {
	cookie, err := r.Cookie(string(e))
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", ErrNoStringValue
	}
	return cookie.Value, nil
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
