package extract

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (singleQuery)("")
	_ StringValueExtractor  = (singleQuery)("")
	_ RequestValueExtractor = (multiQuery)(nil)
	_ StringValueExtractor  = (multiQuery)(nil)
)

// NewQueryValueExtractor is a [Extractor] extractor that pull
// out [url.URL] query values by name from an [http.Request].
func NewQueryValueExtractor(headerNames ...string) (Extractor, error) {
	total := len(headerNames)
	if total == 0 {
		return nil, errors.New("URL query value extractor requires at least one parameter name")
	}
	if err := uniqueNonEmptyValueNames(headerNames); err != nil {
		return nil, err
	}
	if total == 1 {
		return singleQuery(headerNames[0]), nil
	}
	return multiQuery(headerNames), nil
}

type singleQuery string

func (e singleQuery) ExtractRequestValue(vs url.Values, r *http.Request) error {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}

	desired := string(e)
	for name, value := range values {
		if len(value) > 0 {
			if name == desired {
				vs[name] = value
				break
			}
		}
	}
	return nil
}

func (e singleQuery) ExtractStringValue(r *http.Request) (string, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return "", err
	}

	desired := string(e)
	for name, value := range values {
		if last := len(value); last > 0 {
			if name == desired {
				return value[last-1], nil
			}
		}
	}
	return "", NoValueError{desired}
}

type multiQuery []string

func (e multiQuery) ExtractRequestValue(vs url.Values, r *http.Request) error {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}

	for _, desired := range e {
		for name, value := range values {
			if len(value) > 0 {
				if name == desired {
					vs[name] = value
					break
				}
			}
		}
	}
	return nil
}

func (e multiQuery) ExtractStringValue(r *http.Request) (string, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return "", err
	}

	for _, desired := range e {
		for name, value := range values {
			if last := len(value); last > 0 {
				if name == desired {
					return value[last-1], nil
				}
			}
		}
	}
	return "", NoValueError(e)
}
