package extract

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	_ RequestValueExtractor = (singleCookie)("")
	_ StringValueExtractor  = (singleCookie)("")
	_ RequestValueExtractor = (multiCookie)(nil)
	_ StringValueExtractor  = (multiCookie)(nil)
)

// NewCookieValueExtractor is a [Extractor] extractor that pull
// out [http.Cookie] values by name from an [http.Request].
func NewCookieValueExtractor(names ...string) (Extractor, error) {
	total := len(names)
	if total == 0 {
		return nil, errors.New("HTTP cookie value extractor requires at least one cookie name")
	}
	if err := uniqueNonEmptyValueNames(names); err != nil {
		return nil, err
	}
	if total == 1 {
		return singleCookie(names[0]), nil
	}
	return multiCookie(names), nil
}

type singleCookie string

func (e singleCookie) ExtractRequestValue(vs url.Values, r *http.Request) error {
	desired := string(e)
	for _, cookie := range r.Cookies() {
		if cookie.Name == desired && len(cookie.Value) > 0 {
			vs[desired] = []string{cookie.Value}
			break
		}
	}
	return nil
}

func (e singleCookie) ExtractStringValue(r *http.Request) (string, error) {
	desired := string(e)
	for _, cookie := range r.Cookies() {
		if cookie.Name == desired && len(cookie.Value) > 0 {
			return cookie.Value, nil
		}
	}
	return "", ErrNoStringValue
}

type multiCookie []string

func (e multiCookie) ExtractRequestValue(vs url.Values, r *http.Request) error {
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

func (e multiCookie) ExtractStringValue(r *http.Request) (string, error) {
	for _, desired := range e {
		for _, cookie := range r.Cookies() {
			if cookie.Name == desired && len(cookie.Value) > 0 {
				return cookie.Value, nil
			}
		}
	}
	return "", ErrNoStringValue
}
