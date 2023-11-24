package htadaptor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/dkotik/htadaptor/schema"
)

const oneMB = 1 << 20

var structSchema = schema.NewDecoder()

type Decoder interface {
	Decode(any, *http.Request) error
}

// StructDecoder provides configurable HTTP form body decoding that depends on Gorilla's schema reflection package.
type StructDecoder struct {
	// TODO: all the chaching could be ripped out `schema` package as schema can expect only one type of Request struct? or that would needlessly clone schema decoders?
	// schema      *schema.Decoder
	readLimit   int64
	memoryLimit int64
	extractor   RequestValueExtractor
}

func NewStructDecoder(withOptions ...StructDecoderOption) (_ *StructDecoder, err error) {
	o := &structDecoderOptions{}
	if err = WithStructDecoderOptions(append(withOptions,
		func(o *structDecoderOptions) (err error) {
			if o.ReadLimit == 0 {
				if err = WithDefaultDecoderReadLimitOf10MB()(o); err != nil {
					return err
				}
			}
			if o.MemoryLimit == 0 {
				if err = WithDefaultDecoderMemoryLimitOfOneThirdOfReadLimit()(o); err != nil {
					return err
				}
			}
			return nil
		},
	)...)(o); err != nil {
		return nil, fmt.Errorf("cannot initialize a decoder: %w", err)
	}
	return &StructDecoder{
		// schema:      o.Schema,
		readLimit:   o.ReadLimit,
		memoryLimit: o.MemoryLimit,
		extractor:   JoinRequestValueExtractors(o.Extractors...),
	}, nil
}

func (d *StructDecoder) Decode(v any, r *http.Request) (err error) {
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

type structDecoderOptions struct {
	// Schema      *schema.Decoder
	ReadLimit   int64
	MemoryLimit int64
	Extractors  []RequestValueExtractor
}

type StructDecoderOption func(*structDecoderOptions) error

func WithStructDecoderOptions(withOptions ...StructDecoderOption) StructDecoderOption {
	return func(o *structDecoderOptions) (err error) {
		for _, option := range withOptions {
			if option == nil {
				return errors.New("cannot use a <nil> option")
			}
			if err = option(o); err != nil {
				return err
			}
		}
		return nil
	}
}

func WithDecoderReadLimit(upto int64) StructDecoderOption {
	return func(o *structDecoderOptions) error {
		if upto < 1 {
			return errors.New("read limit cannot be less than 1")
		}
		if upto > oneMB*1_000_000 { // math.MaxInt64
			return errors.New("read limit is too large")
		}
		if o.ReadLimit != 0 {
			return errors.New("read limit is already set")
		}
		o.ReadLimit = upto
		return nil
	}
}

func WithDefaultDecoderReadLimitOf10MB() StructDecoderOption {
	return WithDecoderReadLimit(oneMB * 10)
}

func WithDecoderMemoryLimit(upto int64) StructDecoderOption {
	return func(o *structDecoderOptions) error {
		if upto < 1 {
			return errors.New("memory limit cannot be less than 1")
		}
		if upto > oneMB*10_000 { // math.MaxInt64
			return errors.New("memory limit is too large")
		}
		if o.MemoryLimit != 0 {
			return errors.New("memory limit is already set")
		}
		o.MemoryLimit = upto
		return nil
	}
}

func WithDefaultDecoderMemoryLimitOfOneThirdOfReadLimit() StructDecoderOption {
	return func(o *structDecoderOptions) error {
		if o.ReadLimit == 0 {
			return errors.New("read limit is required before default memory limit maybe set")
		}
		return WithDecoderMemoryLimit(o.ReadLimit/3 + 1)(o)
	}
}

func WithDecoderExtractors(exs ...RequestValueExtractor) StructDecoderOption {
	return func(o *structDecoderOptions) error {
		if len(exs) < 1 {
			return errors.New("at least one request value extractor is required")
		}
		for _, ex := range exs {
			if ex == nil {
				return errors.New("cannot use a <nil> request value extractor")
			}
		}
		o.Extractors = append(o.Extractors, exs...)
		return nil
	}
}

func WithDecoderQueryValues(names ...string) StructDecoderOption {
	return func(o *structDecoderOptions) error {
		ex, err := NewQueryValueExtractor(names...)
		if err != nil {
			return fmt.Errorf("failed to initialize query value extractor: %w", err)
		}
		return WithDecoderExtractors(ex)(o)
	}
}

func WithDecoderHeaderValues(names ...string) StructDecoderOption {
	return func(o *structDecoderOptions) error {
		ex, err := NewHeaderValueExtractor(names...)
		if err != nil {
			return fmt.Errorf("failed to initialize header value extractor: %w", err)
		}
		return WithDecoderExtractors(ex)(o)
	}
}

func WithDecoderPathValues(names ...string) StructDecoderOption {
	return func(o *structDecoderOptions) error {
		// TODO: add WithPathValues when 1.22 routing comes out.
		return errors.New("required Go version 1.22")
		// ex, err := NewPathValueExtractor(names...)
		// if err != nil {
		// 	return fmt.Errorf("failed to initialize path value extractor: %w", err)
		// }
		// return WithDecoderExtractors(ex)(o)
	}
}
