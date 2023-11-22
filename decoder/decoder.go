/*
Package decoder provides configurable HTTP form body decoding that depends on Gorilla's schema package which does not contaminate objects with URL query parameters unlike the standard library.
*/
package decoder

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/dkotik/htadaptor/decoder/schema"
)

var structSchema = schema.NewDecoder()

type Extractor func(*http.Request) (url.Values, error)

type Decoder struct {
	// schema      *schema.Decoder
	readLimit   int64
	memoryLimit int64
	extractors  []Extractor
}

func (d *Decoder) Decode(v any, r *http.Request) error {
	values := make(url.Values)
	for _, extractor := range d.extractors {
		more, err := extractor(r)
		if err != nil {
			return err
		}
		for key, valueSet := range more {
			values[key] = valueSet
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

func New(withOptions ...Option) (_ *Decoder, err error) {
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
		extractors:  o.Extractors,
	}, nil
}

// TODO: all the chaching could be ripped out `schema` package as schema can expect only one type of Request struct?
// TODO: Add QueryValues and PathValues (from URL body) to produce url.Values and feed them to Gorilla scheme Decoder when 1.22 comes out.
// id := r.PathValue("id")

func ParseMultiPartBody(r io.Reader, boundary string, memoryLimit int64) (url.Values, error) {
	form, err := multipart.NewReader(r, boundary).ReadForm(memoryLimit)
	if err != nil {
		return nil, err
	}
	values := make(url.Values)
	for k, v := range form.Value {
		values[k] = append(values[k], v...)
	}
	// TODO: attachments can be injected using form.File: map[string][]*FileHeader.
	return values, nil
}

/*
if r.Body == nil {
  err = errors.New("missing form body")
  return
}
ct := r.Header.Get("Content-Type")
// RFC 7231, section 3.1.1.5 - empty type
//   MAY be treated as application/octet-stream
if ct == "" {
  ct = "application/octet-stream"
}
ct, params, err = mime.ParseMediaType(ct) // params includes boundary
switch {
case ct == "application/x-www-form-urlencoded":
  var reader io.Reader = r.Body
  maxFormSize := int64(1<<63 - 1)
  if _, ok := r.Body.(*maxBytesReader); !ok {
    maxFormSize = int64(10 << 20) // 10 MB is a lot of text.
    reader = io.LimitReader(r.Body, maxFormSize+1)
  }
  b, e := io.ReadAll(reader)
  if e != nil {
    if err == nil {
      err = e
    }
    break
  }
  if int64(len(b)) > maxFormSize {
    err = errors.New("http: POST too large")
    return
  }
  vs, e = url.ParseQuery(string(b))
  if err == nil {
    err = e
  }
case ct == "multipart/form-data": // or "multipart/mixed"

    // d, params, err := mime.ParseMediaType(v)
    // if err != nil || !(d == "multipart/form-data" || allowMixed && d == "multipart/mixed") {
    // 	return nil, ErrNotMultipart
    // }
    // boundary, ok := params["boundary"]
    // if !ok {
    // 	return nil, ErrMissingBoundary
    // }
    // return multipart.NewReader(r.Body, boundary), nil

    mr, err := r.multipartReader(false) // bool is allowMixed
    if err != nil {
      return err
    }

    f, err := mr.ReadForm(maxMemory)
    if err != nil {
      return err
    }

    if r.PostForm == nil {
      r.PostForm = make(url.Values)
    }
    for k, v := range f.Value {
      r.Form[k] = append(r.Form[k], v...)
      // r.PostForm should also be populated. See Issue 9305.
      r.PostForm[k] = append(r.PostForm[k], v...)
    }
case "text/json":
  // ...
}
*/
