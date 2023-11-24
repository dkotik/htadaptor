package htadaptor

import "net/http"

type Encoder interface {
	Encode(http.ResponseWriter, any) error
}

type EncoderFunc func(http.ResponseWriter, any) error

func (f EncoderFunc) Encode(w http.ResponseWriter, v any) error {
	return f(w, v)
}
