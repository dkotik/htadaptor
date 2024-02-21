package session

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
)

func newCodec() *securecookie.SecureCookie {
	return securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)
}

type cookieEncoder struct {
	name string
	path string

	mu      *sync.Mutex
	current *securecookie.SecureCookie
}

func (c *cookieEncoder) Encode(
	w http.ResponseWriter,
	r *http.Request,
	data any,
	expires time.Time,
) error {
	c.mu.Lock()
	// TODO: enforce MaxAge on data, otherwise
	// sessions are eternal as long as they are written to regularly.
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
		Expires:  expires,
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
		if errors.Is(http.ErrNoCookie, err) {
			return nil
		}
		return err
	}

	c.mu.Lock()
	err = securecookie.DecodeMulti(c.name, cookie.Value, data, c.current, c.previous)
	c.mu.Unlock()
	if err == nil {
		return nil
	}

	decodingError, ok := err.(interface {
		IsDecode() bool
	})
	// var decodingError securecookie.Error
	// var decodingError securecookie.MultiError
	// if !errors.As(decodingError, err) || !decodingError.IsDecode() {
	if ok && decodingError.IsDecode() {
		// panic(decodingError.IsDecode())
		return nil
	}

	// log.Printf("decoding %x %x", &c.current, &c.previous)
	return err
}
