package session

import (
	"net/http"
	"sync"

	"github.com/gorilla/securecookie"
)

func newCodec() *securecookie.SecureCookie {
	return securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)
}

type cookieEncoder struct {
	name   string
	path   string
	maxAge int

	mu      *sync.Mutex
	current *securecookie.SecureCookie
}

func (c *cookieEncoder) Encode(
	w http.ResponseWriter,
	r *http.Request,
	data any,
) error {
	c.mu.Lock()
	encoded, err := c.current.Encode(c.name, data)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()

	w.Header().Add("Set-Cookie", (&http.Cookie{
		Name:     c.name,
		Value:    encoded,
		Path:     c.path,
		MaxAge:   c.maxAge,
		SameSite: http.SameSiteStrictMode,
	}).String())
	return nil
}

type cookieDecoder struct {
	name string

	mu       *sync.Mutex
	current  *securecookie.SecureCookie
	previous *securecookie.SecureCookie
}

func (c *cookieDecoder) Decode(
	data any,
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	cookie, err := r.Cookie(c.name)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// log.Printf("decoding %x %x", &c.current, &c.previous)
	return securecookie.DecodeMulti(c.name, cookie.Value, data, c.current, c.previous)
}
