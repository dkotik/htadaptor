package extract

import (
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (*host)(nil)
	_ StringValueExtractor  = (*host)(nil)
)

// NewHostExtractor pulls host name an [http.Request].
func NewHostExtractor() (Extractor, error) {
	return &host{}, nil
}

type host struct{}

func (h *host) ExtractRequestValue(vs url.Values, r *http.Request) error {
	vs["host"] = []string{r.Host}
	return nil
}

func (h *host) ExtractStringValue(r *http.Request) (string, error) {
	if len(r.Host) > 0 {
		return r.Host, nil
	}
	return "", NoValueError{"host"}
}
