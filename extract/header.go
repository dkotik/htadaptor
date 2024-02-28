package extract

import (
	"errors"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

var (
	_ RequestValueExtractor = (*singleHeader)(nil)
	_ StringValueExtractor  = (*singleHeader)(nil)
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

	r := strings.NewReplacer("-", "")
	associations := make([]Association, total)
	for i, h := range headerNames {
		associations[i].RequestName = h
		associations[i].SchemaName = r.Replace(h)
	}
	if total == 1 {
		return singleHeader(associations[0]), nil
	}
	return multiHeader(associations), nil
}

type singleHeader Association

func (e singleHeader) ExtractRequestValue(vs url.Values, r *http.Request) error {
	valueSet := textproto.MIMEHeader(r.Header)
	// found := r.Header.Values(desired)
	found := valueSet.Values(e.RequestName)
	if len(found) > 0 {
		vs[e.SchemaName] = found
	}
	return nil
}

func (e singleHeader) ExtractStringValue(r *http.Request) (string, error) {
	valueSet := textproto.MIMEHeader(r.Header)
	// found := r.Header.Values(desired)
	found := valueSet.Values(e.RequestName)
	if last := len(found); last > 0 {
		return found[last-1], nil
	}
	return "", ErrNoStringValue
}

type multiHeader []Association

func (e multiHeader) ExtractRequestValue(vs url.Values, r *http.Request) error {
	valueSet := textproto.MIMEHeader(r.Header)
	for _, desired := range e {
		found := valueSet.Values(desired.RequestName)
		if len(found) > 0 {
			vs[desired.SchemaName] = found
		}
	}
	return nil
}

func (e multiHeader) ExtractStringValue(r *http.Request) (string, error) {
	valueSet := textproto.MIMEHeader(r.Header)
	for _, desired := range e {
		found := valueSet.Values(desired.RequestName)
		if last := len(found); last > 0 {
			return found[last-1], nil
		}
	}
	return "", ErrNoStringValue // TODO: change to work with [Association].
}
