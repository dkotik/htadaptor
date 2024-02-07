package extractor

import (
	"net/http"
	"net/url"
)

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
