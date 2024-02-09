package extract

import (
	"net/http"
	"net/url"
)

type CookieValueExtractor []string

func NewCookieValueExtractor(names ...string) (CookieValueExtractor, error) {
	if err := uniqueNonEmptyValueNames(names); err != nil {
		return nil, err
	}
	return CookieValueExtractor(names), nil
}

func (e CookieValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	for _, cookie := range r.Cookies() {
		for _, desired := range e {
			if cookie.Name == desired && len(cookie.Value) > 0 {
				vs[desired] = []string{cookie.Value}
				break
			}
		}
	}
	return nil
}
