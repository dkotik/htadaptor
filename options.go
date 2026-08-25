package htadaptor

import (
	"errors"
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/dkotik/htadaptor/extract"
	"github.com/dkotik/htadaptor/reflectd"
)

type options struct {
	Decoder        Decoder
	DecoderOptions []reflectd.Option
	Encoder        Encoder
	StatusCode     int
	ErrorHandler   ErrorHandler
}

type Option func(*options) error

func WithOptions(withOptions ...Option) Option {
	return func(o *options) (err error) {
		for _, option := range withOptions {
			if option == nil {
				return errors.New("cannot use a <nil> htadaptor option")
			}
			if err = option(o); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithEncoder sets the response encoder for the handler.
// It is often a template-based encoder such as [NewTemplateEncoder]
// or an API view such as [JSONEncoder].
//
// This option overrides any previously set encoder.
func WithEncoder(e Encoder) Option {
	return func(o *options) error {
		if e == nil {
			return errors.New("cannot use a <nil> response encoder")
		}
		o.Encoder = e
		return nil
	}
}

// WithStatusCode applies HTTP status code to the
// encoder set using [WithEncoder] option to override the
// default [http.StatusOK] success code or any previously set
// value.
func WithStatusCode(statusCode int) Option {
	return func(o *options) error {
		if statusCode < 1 {
			return errors.New("invalid status code")
		}
		// status code is applied inside [WithDefaultEncoder]
		o.StatusCode = statusCode
		return nil
	}
}

// WithTemplate is a convenient version of [WithEncoder] option that
// applies [NewTemplateEncoder] to the given template.
//
// This option overrides any previously set encoder.
func WithTemplate(t *template.Template) Option {
	return func(o *options) error {
		if t == nil {
			return errors.New("cannot use a <nil> template")
		}
		return WithEncoder(NewTemplateEncoder(t))(o)
	}
}

func WithDefaultEncoder() Option {
	return func(o *options) error {
		if o.Encoder != nil {
			return nil
		}
		return WithEncoder(JSONEncoder)(o)
	}
}

func WithErrorHandler(h ErrorHandler) Option {
	return func(o *options) error {
		if h == nil {
			return errors.New("cannot use a <nil> error handler")
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
	return func(o *options) (err error) {
		if o.ErrorHandler != nil {
			return nil
		}

		// capture encoder content type
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		if err = o.Encoder.Encode(w, r, http.StatusOK, nil); err != nil {
			return fmt.Errorf("assigned encoder failed to encode <nil>: %w", err)
		}
		contentType, _, err := mime.ParseMediaType(
			w.Result().Header.Get("content-type"))
		if err != nil {
			return fmt.Errorf("unable to parse content type of the encoded response: %w", err)
		}

		switch contentType {
		case "application/json":
			defaultErrorHandlerJSONSetup.Do(func() {
				defaultErrorHandlerJSON = NewErrorHandler(JSONEncoder)
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
		o.DecoderOptions = nil
		o.Decoder = d
		return nil
	}
}

func WithDecoderOptions(withOptions ...reflectd.Option) Option {
	return func(o *options) error {
		o.Decoder = nil
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

// WithSessionValues is a convenience option that adds [reflectd.WithSessionValues] to the decoder options.
func WithSessionValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithSessionValues(names...))
		return nil
	}
}
