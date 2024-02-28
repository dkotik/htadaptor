package extract

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (singlePath)("")
	_ StringValueExtractor  = (singlePath)("")
	_ RequestValueExtractor = (multiPath)(nil)
	_ StringValueExtractor  = (multiPath)(nil)
)

// NewPathValueExtractor is a [Extractor] extractor that pull
// out [url.URL] path values by name from an [http.Request].
func NewPathValueExtractor(names ...string) (Extractor, error) {
	total := len(names)
	if total == 0 {
		return nil, errors.New("URL path value extractor requires at least one path segment name")
	}
	if err := uniqueNonEmptyValueNames(names); err != nil {
		return nil, err
	}
	if total == 1 {
		return singlePath(names[0]), nil
	}
	return multiPath(names), nil
}

type singlePath string

func (e singlePath) ExtractRequestValue(
	vs url.Values,
	r *http.Request,
) error {
	desired := string(e)
	if value := r.PathValue(desired); value != "" {
		vs[desired] = []string{value}
	}
	return nil
}

func (e singlePath) ExtractStringValue(r *http.Request) (string, error) {
	desired := string(e)
	if value := r.PathValue(desired); value != "" {
		return value, nil
	}
	return "", ErrNoStringValue
}

type multiPath []string

func (e multiPath) ExtractRequestValue(
	vs url.Values,
	r *http.Request,
) error {
	for _, desired := range e {
		if value := r.PathValue(desired); value != "" {
			vs[desired] = []string{value}
		}
	}
	return nil
}

func (e multiPath) ExtractStringValue(r *http.Request) (string, error) {
	for _, desired := range e {
		if value := r.PathValue(desired); value != "" {
			return value, nil
		}
	}
	return "", ErrNoStringValue
}
