package extract

import (
	"errors"
	"net/http"
	"net/url"
)

type CookieValueExtractor string

func NewCookieValueExtractor(names ...string) (RequestValueExtractor, error) {
	total := len(names)
	if total == 0 {
		return nil, errors.New("HTTP cookie value extractor requires at least one cookie name")
	}
	if err := uniqueNonEmptyValueNames(names); err != nil {
		return nil, err
	}
	if total == 1 {
		return CookieValueExtractor(names[0]), nil
	}
	extractors := make([]RequestValueExtractor, total)
	for i, name := range names {
		extractors[i] = CookieValueExtractor(name)
	}
	return Join(extractors...), nil
}

func (e CookieValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	desired := string(e)
	for _, cookie := range r.Cookies() {
		if cookie.Name == desired && len(cookie.Value) > 0 {
			vs[desired] = []string{cookie.Value}
			break
		}
	}
	return nil
}
