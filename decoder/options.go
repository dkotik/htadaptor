package decoder

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const oneMB = 1 << 20

type options struct {
	// Schema      *schema.Decoder
	ReadLimit   int64
	MemoryLimit int64
	Extractors  []Extractor
	// PathValues  []string
}

type Option func(*options) error

func WithOptions(withOptions ...Option) Option {
	return func(o *options) (err error) {
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

func WithReadLimit(upto int64) Option {
	return func(o *options) error {
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

func WithDefaultReadLimitOf10MB() Option {
	return WithReadLimit(oneMB * 10)
}

func WithMemoryLimit(upto int64) Option {
	return func(o *options) error {
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

func WithDefaultMemoryLimitOfOneThirdOfReadLimit() Option {
	return func(o *options) error {
		if o.ReadLimit == 0 {
			return errors.New("read limit is required before default memory limit maybe set")
		}
		return WithMemoryLimit(o.ReadLimit/3 + 1)(o)
	}
}

func WithExtractor(e Extractor) Option {
	return func(o *options) error {
		if e == nil {
			return errors.New("cannot use a <nil> extractor")
		}
		o.Extractors = append(o.Extractors, e)
		return nil
	}
}

// TODO: add WithPathValues when 1.22 routing comes out.
func WithQueryValues(names ...string) Option {
	return func(o *options) error {
		if len(names) == 0 {
			return errors.New("provide at least one URL query parameter name")
		}
		found := make(map[string]struct{})
		for _, name := range names {
			if name == "" {
				return errors.New("cannot use an empty URL query parameter name")
			}
			if _, ok := found[name]; ok {
				return fmt.Errorf("URL query parameter %q is listed more than once", name)
			}
			found[name] = struct{}{}
		}
		return WithExtractor(func(r *http.Request) (url.Values, error) {
			values, err := url.ParseQuery(r.URL.RawQuery)
			if err != nil {
				return nil, err
			}
			result := make(url.Values)
			for name, value := range values {
				for _, desired := range names {
					if name == desired {
						result[name] = value
					}
				}
			}
			return result, nil
		})(o)
	}
}

func WithHeaderValues(names ...string) Option {
	return func(o *options) error {
		if len(names) == 0 {
			return errors.New("provide at least one HTTP header name")
		}
		found := make(map[string]struct{})
		for _, name := range names {
			if name == "" {
				return errors.New("cannot use an empty HTTP header name")
			}
			if _, ok := found[name]; ok {
				return fmt.Errorf("HTTP header %q is listed more than once", name)
			}
			found[name] = struct{}{}
		}
		return WithExtractor(func(r *http.Request) (url.Values, error) {
			result := make(url.Values)
			for _, name := range names {
				result[name] = r.Header.Values(name)
			}
			return result, nil
		})(o)
	}
}
