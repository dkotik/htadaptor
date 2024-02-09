package extract

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (*PathValueExtractor)(nil)
	_ StringValueExtractor  = (*PathValueExtractor)(nil)
)

type PathValueExtractor string

func NewPathValueExtractor(names ...string) (RequestValueExtractor, error) {
	total := len(names)
	if total == 0 {
		return nil, errors.New("URL path value extractor requires at least one path segment name")
	}
	if err := uniqueNonEmptyValueNames(names); err != nil {
		return nil, err
	}
	if total == 1 {
		return PathValueExtractor(names[0]), nil
	}
	extractors := make([]RequestValueExtractor, total)
	for i, name := range names {
		extractors[i] = PathValueExtractor(name)
	}
	return Join(extractors...), nil
}

func (e PathValueExtractor) ExtractRequestValue(
	vs url.Values,
	r *http.Request,
) error {
	desired := string(e)
	if value := r.PathValue(desired); value != "" {
		vs[desired] = []string{value}
	}
	return nil
}

// ExtractStringValue satisfies [StringValue] interface.
func (e PathValueExtractor) ExtractStringValue(r *http.Request) (string, error) {
	desired := string(e)
	if value := r.PathValue(desired); value != "" {
		return value, nil
	}
	return "", NoValueError(desired)
}
