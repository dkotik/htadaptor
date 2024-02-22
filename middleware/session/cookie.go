package session

import (
	"net/http"
	"time"
)

// A browser should be able to accept at least 300 cookies with a maximum size of 4096 bytes, as stipulated by RFC 2109 (#6.3), RFC 2965 (#5.3), and RFC 6265.
const MaximumCookieSize = 4096

type CookieCodec interface {
	WriteCookie(http.ResponseWriter, string, time.Time) error
	ReadCookie(*http.Request) string
}

type strictCookieCodec struct {
	Name string
	Path string
}

func NewStrictCookieCodec(name, path string) CookieCodec {
	return &strictCookieCodec{
		Name: name,
		Path: path,
	}
}

func (c *strictCookieCodec) WriteCookie(
	w http.ResponseWriter,
	value string,
	expires time.Time,
) error {
	cookie := (&http.Cookie{
		Name:     c.Name,
		Value:    value,
		Path:     c.Path,
		Expires:  expires,
		SameSite: http.SameSiteStrictMode,
	}).String()
	if len(cookie) > MaximumCookieSize {
		return ErrLargeCookie
	}
	w.Header().Add("Set-Cookie", cookie)
	return nil
}

func (c *strictCookieCodec) ReadCookie(r *http.Request) string {
	cookie, err := r.Cookie(c.Name)
	if err != nil {
		return ""
		// if errors.Is(http.ErrNoCookie, err) {
		//   return nil
		// }
		// return err
	}
	return cookie.Value
}
