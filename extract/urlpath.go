package extract

import (
	"net/http"
	"net/url"
)

type PathValueExtractor []string

func NewPathValueExtractor(names ...string) (PathValueExtractor, error) {
	if err := uniqueNonEmptyValueNames(names); err != nil {
		return nil, err
	}
	return PathValueExtractor(names), nil
}

func (e PathValueExtractor) ExtractRequestValue(
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
