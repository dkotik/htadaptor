package reflectd

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

func (d *Decoder) DecodeJSON(v any, r *http.Request) (err error) {
	if r.Body != nil {
		body, err := io.ReadAll(io.LimitReader(r.Body, d.readLimit))
		err = errors.Join(err, r.Body.Close())
		if err != nil {
			return err
		}
		if len(body) > 0 {
			if err = json.Unmarshal(body, &v); err != nil {
				return err
			}
		}
	}
	values := make(url.Values)
	if err = d.applyExtractors(values, r); err != nil {
		return err
	}
	return structSchema.Decode(v, values)
}
