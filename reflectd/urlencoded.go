package reflectd

import (
	"errors"
	"io"
	"net/http"
	"net/url"
)

func (d *Decoder) DecodeURLEncoded(v any, r *http.Request) (err error) {
	values := make(url.Values)
	if r.Body != nil {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return errors.Join(err, r.Body.Close())
		}
		if err = r.Body.Close(); err != nil {
			return err
		}
		values, err = url.ParseQuery(string(b))
		if err != nil {
			return err
		}
	}

	if err = d.applyExtractors(values, r); err != nil {
		return err
	}
	return structSchema.Decode(v, values)
}
