package reflectd

import (
	"errors"
	"fmt"

	"github.com/dkotik/htadaptor/extractor"
)

type options struct {
	// Schema      *schema.Decoder
	ReadLimit   int64
	MemoryLimit int64
	Extractors  []extractor.RequestValueExtractor
}

// Option configures new [Decoder]s.
type Option func(*options) error

// WithOptions combines several [Option]s into one.
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

// WithReadLimit contrains acceptable [request.Body] size.
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

// WithDefaultReadLimitOf10MB enforces the standard library convention when [WithReadLimit] is not used.
func WithDefaultReadLimitOf10MB() Option {
	return WithReadLimit(oneMB * 10)
}

// WithMemoryLimit constrains system memory usage when decoding mixed multipart forms, which typically include file uploads. When the memory limit is exceeded, the excess data is written to disk. Higher limit speeds up handling file uploads.
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

// WithDefaultMemoryLimitOfOneThirdOfReadLimit contrains memory limit to one third of the read limit when [WithMemoryLimit] option is not used. Use [WithReadLimit] option to adjust the read limit.
func WithDefaultMemoryLimitOfOneThirdOfReadLimit() Option {
	return func(o *options) error {
		if o.ReadLimit == 0 {
			return errors.New("read limit is required before default memory limit maybe set")
		}
		return WithMemoryLimit(o.ReadLimit/3 + 1)(o)
	}
}

// WithExtractors adds [extractor.RequestValueExtractor]s to a [Decoder]. The order of extractors determines their precedence.
func WithExtractors(exs ...extractor.RequestValueExtractor) Option {
	return func(o *options) error {
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

// WithQueryValues adds a [QueryValueExtractor] to a [Decoder].
func WithQueryValues(names ...string) Option {
	return func(o *options) error {
		ex, err := extractor.NewQueryValueExtractor(names...)
		if err != nil {
			return fmt.Errorf("failed to initialize query value extractor: %w", err)
		}
		return WithExtractors(ex)(o)
	}
}

// WithHeaderValues adds a [HeaderValueExtractor] to a [Decoder].
func WithHeaderValues(names ...string) Option {
	return func(o *options) error {
		ex, err := extractor.NewHeaderValueExtractor(names...)
		if err != nil {
			return fmt.Errorf("failed to initialize header value extractor: %w", err)
		}
		return WithExtractors(ex)(o)
	}
}

// WithCookieValues adds a [CookieValueExtractor] to a [Decoder].
func WithCookieValues(names ...string) Option {
	return func(o *options) error {
		ex, err := extractor.NewCookieValueExtractor(names...)
		if err != nil {
			return fmt.Errorf("failed to initialize cookie value extractor: %w", err)
		}
		return WithExtractors(ex)(o)
	}
}

// WithPathValues adds a [PathValueExtractor] to a [Decoder].
func WithPathValues(names ...string) Option {
	return func(o *options) error {
		ex, err := extractor.NewPathValueExtractor(names...)
		if err != nil {
			return fmt.Errorf("failed to initialize path value extractor: %w", err)
		}
		return WithExtractors(ex)(o)
	}
}
