package reflectd

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

func (d *Decoder) DecodeMultiPart(v any, r *http.Request, boundary string) (err error) {
	values := make(url.Values)
	if r.Body != nil {
		form, err := multipart.NewReader(
			io.LimitReader(r.Body, d.readLimit),
			boundary,
		).ReadForm(d.memoryLimit)
		if err != nil {
			return err
		}
		for k, v := range form.Value {
			values[k] = append(values[k], v...)
		}
		// TODO: clean up form when context expires.
		// _ = context.AfterFunc(ctx, func() {
		//   if err := form.Clean(); err != nil {
		//     warn...
		//   }
		// })
		// TODO: attachments can be injected using form.File: map[string][]*FileHeader.
	}
	if err = d.applyExtractors(values, r); err != nil {
		return err
	}
	return structSchema.Decode(v, values)
}
