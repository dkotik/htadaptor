package htadaptor

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"sync"

	"github.com/dkotik/htadaptor/extract"
	"github.com/dkotik/htadaptor/reflectd"
)

// TODO: use Option interface that applies to various options structs?
// TODO: put StringExtractor into common options to remove it as
// parameter from unarystr.go and voidstr.go
// TODO: allow options to overwrite previous sets? or use
// WithDecoderOverride... option set to make override explicit?
type options struct {
	Decoder        Decoder
	DecoderOptions []reflectd.Option
	Encoder        Encoder
	ErrorHandler   ErrorHandler
	Logger         Logger
}

func (o *options) Validate() (err error) {
	if len(o.DecoderOptions) > 0 {
		if o.Decoder != nil {
			return fmt.Errorf("option WithDecoder conflicts with %d decoder options; provide either a prepared decoder or options for preparing one, but not both", len(o.DecoderOptions))
		}
		o.Decoder, err = reflectd.NewDecoder(o.DecoderOptions...)
		if err != nil {
			return err
		}
	}
	return WithOptions(
		WithDefaultEncoder(),
		WithDefaultErrorHandler(),
		WithDefaultLogger(),
	)(o)
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

func WithEncoder(e Encoder) Option {
	return func(o *options) error {
		if e == nil {
			return errors.New("cannot use a <nil> response encoder")
		}
		if o.Encoder != nil {
			return errors.New("response encoder is already set")
		}
		o.Encoder = e
		return nil
	}
}

func WithTemplate(t *template.Template) Option {
	return func(o *options) error {
		if t == nil {
			return errors.New("cannot use a <nil> template")
		}
		return WithEncoder(NewTemplateEncoder(t))(o)
	}
}

var (
	defaultEncoder      Encoder
	defaultEncoderSetup sync.Once
)

func WithDefaultEncoder() Option {
	return func(o *options) error {
		if o.Encoder != nil {
			return nil
		}
		defaultEncoderSetup.Do(func() {
			defaultEncoder = &JSONEncoder{}
		})
		return WithEncoder(defaultEncoder)(o)
	}
}

func WithErrorHandler(h ErrorHandler) Option {
	return func(o *options) error {
		if h == nil {
			return errors.New("cannot use a <nil> error handler")
		}
		if o.ErrorHandler != nil {
			return errors.New("error handler is already set")
		}
		o.ErrorHandler = h
		return nil
	}
}

var (
	defaultErrorHandlerJSON      ErrorHandler
	defaultErrorHandlerJSONSetup sync.Once
	defaultErrorHandlerHTML      ErrorHandler
	defaultErrorHandlerHTMLSetup sync.Once
)

func WithDefaultErrorHandler() Option {
	return func(o *options) error {
		if o.ErrorHandler != nil {
			return nil
		}
		switch o.Encoder.ContentType() {
		case "application/json":
			defaultErrorHandlerJSONSetup.Do(func() {
				defaultErrorHandlerJSON = NewErrorHandler(&JSONEncoder{})
			})
			return WithErrorHandler(defaultErrorHandlerJSON)(o)
		case "text/html":
			defaultErrorHandlerHTMLSetup.Do(func() {
				defaultErrorHandlerHTML = NewErrorHandlerFromTemplate(DefaultErrorTemplate())
				// NewErrorHandler(
				// 	NewTemplateEncoder(DefaultErrorTemplate()))
			})
			return WithErrorHandler(defaultErrorHandlerHTML)(o)
		default:
			return WithErrorHandler(NewErrorHandler(o.Encoder))(o)
		}
	}
}

func WithDecoder(d Decoder) Option {
	return func(o *options) error {
		if d == nil {
			return errors.New("cannot use a <nil> decoder")
		}
		if o.Decoder != nil {
			return errors.New("decoder is already set")
		}
		o.Decoder = d
		return nil
	}
}

func WithDecoderOptions(withOptions ...reflectd.Option) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, withOptions...)
		return nil
	}
}

func WithDefaultDecoder() Option {
	return func(o *options) (err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("cannot initialize default struct decoder: %w", err)
			}
		}()
		d, err := reflectd.NewDecoder()
		if err != nil {
			return err
		}
		return WithDecoder(d)(o)
	}
}

func WithReadLimit(upto int64) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithReadLimit(upto))
		return nil
	}
}

func WithMemoryLimit(upto int64) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithMemoryLimit(upto))
		return nil
	}
}

func WithExtractors(exs ...extract.RequestValueExtractor) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithExtractors(exs...))
		return nil
	}
}

func WithQueryValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithQueryValues(names...))
		return nil
	}
}

func WithHeaderValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithHeaderValues(names...))
		return nil
	}
}

func WithCookieValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithCookieValues(names...))
		return nil
	}
}

func WithPathValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithPathValues(names...))
		return nil
	}
}

func WithSessionValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithSessionValues(names...))
		return nil
	}
}

func WithLogger(l Logger) Option {
	return func(o *options) error {
		if l == nil {
			return errors.New("cannot use a <nil> logger")
		}
		if o.Logger != nil {
			return errors.New("adaptor logger is already set")
		}
		o.Logger = l
		return nil
	}
}

var (
	defaultLogger      Logger
	defaultLoggerSetup sync.Once
)

func WithDefaultLogger() Option {
	return func(o *options) error {
		if o.Logger != nil {
			return nil
		}
		defaultLoggerSetup.Do(func() {
			defaultLogger = &SlogLogger{
				Logger:  slog.Default(),
				Success: slog.LevelDebug,
				Error:   slog.LevelError,
			}
		})
		return WithLogger(defaultLogger)(o)
	}
}
