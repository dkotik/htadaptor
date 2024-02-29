package htadaptor

import "net/http"

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
