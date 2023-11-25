package reflectd

import (
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"net/url"
)

// RequestValueExtractor pulls [url.Values] from an [http.Request]
// in order to provide a [Decoder] with values to populate
// domain request struct with.
type RequestValueExtractor interface {
	ExtractRequestValue(url.Values, *http.Request) error
}

// RequestValueExtractorFunc provides a wrapper for functional
// implementation of a [RequestValueExtractor].
type RequestValueExtractorFunc func(url.Values, *http.Request) error

// ExtractRequestValue satisfies [RequestValueExtractor] interface
// for [RequestValueExtractorFunc].
func (f RequestValueExtractorFunc) ExtractRequestValue(vs url.Values, r *http.Request) error {
	return f(vs, r)
}

// JoinRequestValueExtractors unites several extractors into one.
// Returns <nil> if no extracots are given.
func JoinRequestValueExtractors(exs ...RequestValueExtractor) RequestValueExtractor {
	switch len(exs) {
	case 0:
		return nil
	case 1:
		return exs[0]
	default:
		return RequestValueExtractorFunc(
			func(vs url.Values, r *http.Request) (err error) {
				for _, extractor := range exs {
					if err = extractor.ExtractRequestValue(vs, r); err != nil {
						return err
					}
				}
				return nil
			},
		)
	}
}

func uniqueNonEmptyValueNames(names []string) error {
	if len(names) == 0 {
		return errors.New("provide at least one value name")
	}
	found := make(map[string]struct{})
	for _, name := range names {
		if name == "" {
			return errors.New("cannot use an empty value name")
		}
		if _, ok := found[name]; ok {
			return fmt.Errorf("value name %q occurs more than once", name)
		}
		found[name] = struct{}{}
	}
	return nil
}

type HeaderValueExtractor struct {
	names []string
}

func NewHeaderValueExtractor(headerNames ...string) (*HeaderValueExtractor, error) {
	if err := uniqueNonEmptyValueNames(headerNames); err != nil {
		return nil, err
	}
	return &HeaderValueExtractor{names: headerNames}, nil
}

func (e *HeaderValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	valueSet := textproto.MIMEHeader(r.Header)
	for _, desired := range e.names {
		// found := r.Header.Values(desired)
		found := valueSet.Values(desired)
		if len(found) > 0 {
			vs[desired] = found
		}
	}
	return nil
}

type CookieValueExtractor struct {
	names []string
}

func NewCookieValueExtractor(names ...string) (*CookieValueExtractor, error) {
	if err := uniqueNonEmptyValueNames(names); err != nil {
		return nil, err
	}
	return &CookieValueExtractor{names: names}, nil
}

func (e *CookieValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	for _, cookie := range r.Cookies() {
		for _, desired := range e.names {
			if cookie.Name == desired && len(cookie.Value) > 0 {
				vs[desired] = []string{cookie.Value}
				break
			}
		}
	}
	return nil
}

type QueryValueExtractor struct {
	names []string
}

func NewQueryValueExtractor(headerNames ...string) (*QueryValueExtractor, error) {
	if err := uniqueNonEmptyValueNames(headerNames); err != nil {
		return nil, err
	}
	return &QueryValueExtractor{names: headerNames}, nil
}

func (e *QueryValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}
	for name, value := range values {
		if len(value) > 0 {
			for _, desired := range e.names {
				if name == desired {
					vs[name] = value
					break
				}
			}
		}
	}
	return nil
}

// TODO: Add PathValues (from URL body) to produce url.Values and feed them to Gorilla scheme Decoder when 1.22 comes out.
// id := r.PathValue("id")
