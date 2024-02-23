package extract

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (address)("")
	_ StringValueExtractor  = (address)("")
)

// NewRemoteAddressExtractor pulls host name an [http.Request].
func NewRemoteAddressExtractor(fieldName string) (Extractor, error) {
	if len(fieldName) < 1 {
		return nil, errors.New("field name is required")
	}
	return address(fieldName), nil
}

type address string

func (a address) ExtractRequestValue(vs url.Values, r *http.Request) error {
	vs[string(a)] = []string{r.RemoteAddr}
	return nil
}

func (a address) ExtractStringValue(r *http.Request) (string, error) {
	if len(r.RemoteAddr) > 0 {
		return r.Host, nil
	}
	return "", NoValueError{string(a)}
}
