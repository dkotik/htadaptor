package extract

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (CookieValueExtractor)(nil)
	_ StringValueExtractor  = (CookieValueExtractor)(nil)
)

// CookieValueExtractor is a [RequestValueExtractor] extractor that pull
// out [http.Cookie] values by name from an [http.Request].
type CookieValueExtractor []string

func NewCookieValueExtractor(names ...string) (CookieValueExtractor, error) {
	total := len(names)
	if total == 0 {
		return nil, errors.New("HTTP cookie value extractor requires at least one cookie name")
	}
	if err := uniqueNonEmptyValueNames(names); err != nil {
		return nil, err
	}
	return CookieValueExtractor(names), nil
}

// ExtractRequestValue satisfies [RequestValueExtractor] interface.
func (e CookieValueExtractor) ExtractRequestValue(vs url.Values, r *http.Request) error {
	for _, desired := range e {
		for _, cookie := range r.Cookies() {
			if cookie.Name == desired && len(cookie.Value) > 0 {
				vs[desired] = []string{cookie.Value}
				break
			}
		}
	}
	return nil
}

// ExtractStringValue satisfies [StringValue] interface.
func (e CookieValueExtractor) ExtractStringValue(r *http.Request) (string, error) {
	for _, desired := range e {
		for _, cookie := range r.Cookies() {
			if cookie.Name == desired && len(cookie.Value) > 0 {
				return cookie.Value, nil
			}
		}
	}
	return "", NoValueError(e[0])
}
