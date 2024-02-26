package extract

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (methodExtractor)("")
	_ StringValueExtractor  = (methodExtractor)("")
)

// NewMethodExtractor pulls request method from an [http.Request].
func NewMethodExtractor(fieldName string) (Extractor, error) {
	if len(fieldName) < 1 {
		return nil, errors.New("field name is required for method extraction")
	}
	return methodExtractor(fieldName), nil
}

type methodExtractor string

func (m methodExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	method := r.Method
	if len(method) > 0 {
		vs[string(m)] = []string{method}
	}
	return nil
}

func (m methodExtractor) ExtractStringValue(r *http.Request) (string, error) {
	method := r.Method
	if len(method) > 0 {
		return r.Host, nil
	}
	return "", NoValueError{string(m)}
}
