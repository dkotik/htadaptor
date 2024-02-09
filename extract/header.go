package extract

import (
	"errors"
	"net/http"
	"net/textproto"
	"net/url"
)

var (
	_ RequestValueExtractor = (singleHeader)("")
	_ StringValueExtractor  = (singleHeader)("")
	_ RequestValueExtractor = (multiHeader)(nil)
	_ StringValueExtractor  = (multiHeader)(nil)
)

// NewHeaderValueExtractor builds an [Extractor] that pulls
// out [http.Header] values by name from an [http.Request].
func NewHeaderValueExtractor(headerNames ...string) (Extractor, error) {
	total := len(headerNames)
	if total == 0 {
		return nil, errors.New("HTTP header value extractor requires at least one header name")
	}
	if err := uniqueNonEmptyValueNames(headerNames); err != nil {
		return nil, err
	}
	if total == 1 {
		return singleHeader(headerNames[0]), nil
	}
	return multiHeader(headerNames), nil
}

type singleHeader string

func (e singleHeader) ExtractRequestValue(vs url.Values, r *http.Request) error {
	desired := string(e)
	valueSet := textproto.MIMEHeader(r.Header)
	// found := r.Header.Values(desired)
	found := valueSet.Values(desired)
	if len(found) > 0 {
		vs[desired] = found
	}
	return nil
}

func (e singleHeader) ExtractStringValue(r *http.Request) (string, error) {
	desired := string(e)
	valueSet := textproto.MIMEHeader(r.Header)
	// found := r.Header.Values(desired)
	found := valueSet.Values(desired)
	if last := len(found); last > 0 {
		return found[last-1], nil
	}
	return "", NoValueError{desired}
}

type multiHeader []string

func (e multiHeader) ExtractRequestValue(vs url.Values, r *http.Request) error {
	valueSet := textproto.MIMEHeader(r.Header)
	for _, desired := range e {
		found := valueSet.Values(desired)
		if len(found) > 0 {
			vs[desired] = found
		}
	}
	return nil
}

func (e multiHeader) ExtractStringValue(r *http.Request) (string, error) {
	valueSet := textproto.MIMEHeader(r.Header)
	for _, desired := range e {
		found := valueSet.Values(desired)
		if last := len(found); last > 0 {
			return found[last-1], nil
		}
	}
	return "", NoValueError(e)
}
