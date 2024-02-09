package extract

import (
	"errors"
	"net/http"
	"net/textproto"
	"net/url"
)

var (
	_ RequestValueExtractor = (*HeaderValueExtractor)(nil)
	_ StringValueExtractor  = (*HeaderValueExtractor)(nil)
)

type HeaderValueExtractor string

func NewHeaderValueExtractor(headerNames ...string) (RequestValueExtractor, error) {
	total := len(headerNames)
	if total == 0 {
		return nil, errors.New("HTTP header value extractor requires at least one header name")
	}
	if err := uniqueNonEmptyValueNames(headerNames); err != nil {
		return nil, err
	}
	if total == 1 {
		return HeaderValueExtractor(headerNames[0]), nil
	}
	extractors := make([]RequestValueExtractor, total)
	for i, name := range headerNames {
		extractors[i] = HeaderValueExtractor(name)
	}
	return Join(extractors...), nil
}

func (e HeaderValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	desired := string(e)
	valueSet := textproto.MIMEHeader(r.Header)
	// found := r.Header.Values(desired)
	found := valueSet.Values(desired)
	if len(found) > 0 {
		vs[desired] = found
	}
	return nil
}

// ExtractStringValue satisfies [StringValue] interface.
func (e HeaderValueExtractor) ExtractStringValue(r *http.Request) (string, error) {
	desired := string(e)
	valueSet := textproto.MIMEHeader(r.Header)
	// found := r.Header.Values(desired)
	found := valueSet.Values(desired)
	if last := len(found); last > 0 {
		return found[last-1], nil
	}
	return "", NoValueError(desired)
}
