package htadaptor

import (
	"fmt"
	"net/http"
)

var (
	_ Encoder = (*temporaryRedirectEncoder)(nil)
	_ Encoder = (*permanentRedirectEncoder)(nil)
)

type temporaryRedirectEncoder struct{}

// NewTemporaryRedirectEncoder redirects the HTTP client
// to the location returned by a domain call
// using [http.StatusTemporaryRedirect] status.
//
// If domain call does not return a string, returns an error.
func NewTemporaryRedirectEncoder() Encoder {
	return &temporaryRedirectEncoder{}
}

func (t *temporaryRedirectEncoder) ContentType() string {
	return "text/html"
}

func (t *temporaryRedirectEncoder) Encode(
	w http.ResponseWriter,
	r *http.Request,
	v any,
) error {
	location, ok := v.(string)
	if !ok {
		return fmt.Errorf("redirection encoder received \"%T\" value instead of a string", v)
	}
	http.Redirect(w, r, location, http.StatusTemporaryRedirect)
	return nil
}

type permanentRedirectEncoder struct{}

// NewPermanentRedirectEncoder redirects the HTTP client
// to the location returned by a domain call
// using [http.StatusPermanentRedirect] status.
//
// If domain call does not return a string, returns an error.
func NewPermanentRedirectEncoder() Encoder {
	return &permanentRedirectEncoder{}
}

func (t *permanentRedirectEncoder) ContentType() string {
	return "text/html"
}

func (t *permanentRedirectEncoder) Encode(
	w http.ResponseWriter,
	r *http.Request,
	v any,
) error {
	location, ok := v.(string)
	if !ok {
		return fmt.Errorf("redirection encoder received \"%T\" value instead of a string", v)
	}
	http.Redirect(w, r, location, http.StatusPermanentRedirect)
	return nil
}

type temporaryRedirect string

func (t temporaryRedirect) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	http.Redirect(w, r, string(t), http.StatusTemporaryRedirect)
}

func NewTemporaryRedirect(to string) http.Handler {
	return temporaryRedirect(to)
}

type permanentRedirect string

func (p permanentRedirect) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	http.Redirect(w, r, string(p), http.StatusPermanentRedirect)
}

func NewPermanentRedirect(to string) http.Handler {
	return permanentRedirect(to)
}
