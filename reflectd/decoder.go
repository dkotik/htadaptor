/*
Package reflectd provides configurable HTTP form body to struct decoder that depends on Gorilla's schema reflection package.
*/
package reflectd

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"

	"github.com/dkotik/htadaptor/extract"
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
	extractors  []extract.RequestValueExtractor
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
			if !extract.AreSessionExtractorsLast(o.Extractors...) {
				return errors.New("security failure: all session value extractors must be at the end of the list to prevent other kinds of extractors from overriding their trusted values even when nested")
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
		extractors:  o.Extractors,
	}, nil
}

func (d *Decoder) applyExtractors(values url.Values, r *http.Request) (err error) {
	for _, extractor := range d.extractors {
		if err = extractor.ExtractRequestValue(values, r); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) Decode(v any, r *http.Request) (err error) {
	ct := r.Header.Get("Content-Type")
	switch ct {
	case "application/x-www-form-urlencoded":
		return d.DecodeURLEncoded(v, r)
	case "application/json":
		return d.DecodeJSON(v, r)
	}

	ct, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}
	switch ct {
	case "multipart/form-data", "multipart/mixed":
		boundary, ok := params[`boundary`]
		if !ok {
			return http.ErrMissingBoundary
		}
		return d.DecodeMultiPart(v, r, boundary)
	default:
		return extract.ErrUnsupportedMediaType
	}
}

// func mergeURLValues(b, a url.Values) {
// 	for key, valueSet := range a {
// 		b[key] = valueSet
// 	}
// }
