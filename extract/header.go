package extract

import (
	"net/http"
	"net/textproto"
	"net/url"
)

type HeaderValueExtractor []string

func NewHeaderValueExtractor(headerNames ...string) (HeaderValueExtractor, error) {
	if err := uniqueNonEmptyValueNames(headerNames); err != nil {
		return nil, err
	}
	return HeaderValueExtractor(headerNames), nil
}

func (e HeaderValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	valueSet := textproto.MIMEHeader(r.Header)
	for _, desired := range e {
		// found := r.Header.Values(desired)
		found := valueSet.Values(desired)
		if len(found) > 0 {
			vs[desired] = found
		}
	}
	return nil
}
