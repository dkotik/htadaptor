/*
Package reflectd provides configurable HTTP form body to struct decoder that depends on Gorilla's schema reflection package.
*/
package reflectd

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/dkotik/htadaptor/reflectd/schema"
)

const oneMB = 1 << 20

var structSchema = schema.NewDecoder()

// Decoder provides configurable HTTP form body decoding that depends on Gorilla's schema reflection package.
type Decoder struct {
	// TODO: all the chaching could be ripped out `schema` package as schema can expect only one type of Request struct? or that would needlessly clone schema decoders?
	// schema      *schema.Decoder
	readLimit   int64
	memoryLimit int64
	extractor   RequestValueExtractor
}

func NewDecoder(withOptions ...Option) (_ *Decoder, err error) {
	o := &options{}
	if err = WithOptions(append(withOptions,
		func(o *options) (err error) {
			if o.ReadLimit == 0 {
				if err = WithDefaultReadLimitOf10MB()(o); err != nil {
					return err
				}
			}
			if o.MemoryLimit == 0 {
				if err = WithDefaultMemoryLimitOfOneThirdOfReadLimit()(o); err != nil {
					return err
				}
			}
			return nil
		},
	)...)(o); err != nil {
		return nil, fmt.Errorf("cannot initialize a decoder: %w", err)
	}
	return &Decoder{
		// schema:      o.Schema,
		readLimit:   o.ReadLimit,
		memoryLimit: o.MemoryLimit,
		extractor:   JoinRequestValueExtractors(o.Extractors...),
	}, nil
}

func (d *Decoder) Decode(v any, r *http.Request) (err error) {
	values := make(url.Values)
	if d.extractor != nil {
		if err = d.extractor.ExtractRequestValue(values, r); err != nil {
			return err
		}
	}

	if r.Body == nil {
		return structSchema.Decode(v, values) // body is empty
	}
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return structSchema.Decode(v, values) // unspecified content type
	}
	ct, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	lr := io.LimitReader(r.Body, d.readLimit)

	switch ct {
	case "application/x-www-form-urlencoded":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		more, err := url.ParseQuery(string(b))
		if err != nil {
			return err
		}
		mergeURLValues(values, more)
		return structSchema.Decode(v, values)
	case "multipart/form-data":
		fallthrough
	case "multipart/mixed":
		boundary, ok := params[`boundary`]
		if !ok {
			return http.ErrMissingBoundary
		}
		// TODO: use mime.Form.Files to inject files
		// TODO: mime.Form has Clear method that removes temp files?
		// TODO: mime.Form could be thrown into a GC channel for
		// batch processing.
		// TODO: or can just start a context-waiting goRoutine.
		more, err := ParseMultiPartBody(lr, boundary, d.memoryLimit)
		if err != nil {
			return err
		}
		mergeURLValues(values, more)
		return structSchema.Decode(v, values)
	case "application/json":
		if len(values) > 0 {
			if err = structSchema.Decode(v, values); err != nil {
				return err
			}
		}
		return json.NewDecoder(lr).Decode(&v)
	default:
		return fmt.Errorf("content type %q is not supported", ct)
	}
}

func ParseMultiPartBody(r io.Reader, boundary string, memoryLimit int64) (url.Values, error) {
	form, err := multipart.NewReader(r, boundary).ReadForm(memoryLimit)
	if err != nil {
		return nil, err
	}
	values := make(url.Values)
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
	return values, nil
}

func mergeURLValues(b, a url.Values) {
	for key, valueSet := range a {
		b[key] = valueSet
	}
}
