package extract

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (agent)("")
	_ StringValueExtractor  = (agent)("")
)

// NewUserAgentExtractor pulls host name from an [http.Request].
func NewUserAgentExtractor(fieldName string) (Extractor, error) {
	if len(fieldName) < 1 {
		return nil, errors.New("field name is required")
	}
	return agent(fieldName), nil
}

type agent string

func (a agent) ExtractRequestValue(vs url.Values, r *http.Request) error {
	agent := r.UserAgent()
	if len(agent) > 0 {
		vs[string(a)] = []string{agent}
	}
	return nil
}

func (a agent) ExtractStringValue(r *http.Request) (string, error) {
	agent := r.UserAgent()
	if len(agent) > 0 {
		return r.Host, nil
	}
	return "", ErrNoStringValue
}
