package extract

import (
	"errors"
	"net/http"
	"net/url"
)

type QueryValueExtractor []string

func NewQueryValueExtractor(headerNames ...string) (QueryValueExtractor, error) {
	if len(headerNames) == 0 {
		return nil, errors.New("URL query value extractor requires at least one parameter name")
	}
	if err := uniqueNonEmptyValueNames(headerNames); err != nil {
		return nil, err
	}
	return QueryValueExtractor(headerNames), nil
}

func (e QueryValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}
	for name, value := range values {
		if len(value) > 0 {
			for _, desired := range e {
				if name == desired {
					vs[name] = value
					break
				}
			}
		}
	}
	return nil
}
