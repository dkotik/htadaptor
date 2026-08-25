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
	Middleware     []Middleware
}

// Option intializes the [Adaptor] with defaults
// for [AdaptFunc], [AdaptNullaryFunc], and [AdaptVoidFunc] handlers.
type Option func(*options) error

// WithOptions groups multiple [Option]s into a sequence.
// It is a helper function for creating reusable configuration
// sets.
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

// WithDefaultEncoder sets the default encoder to [JSONEncoder]
// if no other encoder had been set.
func WithDefaultEncoder() Option {
	return func(o *options) error {
		if o.Encoder != nil {
			return nil
		}
		return WithEncoder(JSONEncoder)(o)
	}
}

// WithErrorHandler sets the request error handler.
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

// WithDefaultErrorHandler sets the default error handler
// matching the content type of the encoder if no other
// error handler had been set.
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

// WithUnsafeDecoder sets the decoder for the [Adaptor].
// The decoder is used to decode the request body into a struct.
// If no decoder is set, the default struct decoder is used.
//
// This option is mutually exclusive with [WithDecoderOptions].
// When the decoder is set, all the previously
// decoder options are discarded.
//
// Use this option with caution, as it bypasses the default
// body read limit and parsing memory limit.
func WithUnsafeDecoder(d Decoder) Option {
	return func(o *options) error {
		if d == nil {
			return errors.New("cannot use a <nil> decoder")
		}
		o.DecoderOptions = nil
		o.Decoder = d
		return nil
	}
}

// WithDecoderOptions modifies the default reflection-based
// struct decoder.
//
// This option is mutually exclusive with [WithUnsafeDecoder].
// When the decoder options are set, the previously set
// decoder is discarded.
func WithDecoderOptions(withOptions ...reflectd.Option) Option {
	return func(o *options) error {
		for _, opt := range withOptions {
			if opt == nil {
				return errors.New("cannot use a <nil> decoder option")
			}
		}
		o.Decoder = nil
		o.DecoderOptions = append(o.DecoderOptions, withOptions...)
		return nil
	}
}

// WithDefaultDecoder uses a reflection-based struct decoder.
// Any options set via [WithDecoderOptions] are applied to
// the decoder.
func WithDefaultDecoder() Option {
	return func(o *options) (err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("cannot initialize default struct decoder: %w", err)
			}
		}()
		d, err := reflectd.NewDecoder(o.DecoderOptions...)
		if err != nil {
			return err
		}
		return WithUnsafeDecoder(d)(o)
	}
}

// WithReadLimit sets the read limit for the default decoder.
// This option is only applicable when using the default decoder.
func WithReadLimit(upto int64) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithReadLimit(upto))
		return nil
	}
}

// WithMemoryLimit sets the memory limit for the default decoder.
// This option is only applicable when using the default decoder.
func WithMemoryLimit(upto int64) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithMemoryLimit(upto))
		return nil
	}
}

// WithExtractors sets the extractors for the default decoder.
// This option is only applicable when using the default decoder.
func WithExtractors(exs ...extract.RequestValueExtractor) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithExtractors(exs...))
		return nil
	}
}

// WithQueryValues sets the query values for the default decoder.
// This option is only applicable when using the default decoder.
func WithQueryValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithQueryValues(names...))
		return nil
	}
}

// WithHeaderValues sets the header values for the default decoder.
// This option is only applicable when using the default decoder.
func WithHeaderValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithHeaderValues(names...))
		return nil
	}
}

// WithCookieValues sets the cookie values for the default decoder.
// This option is only applicable when using the default decoder.
func WithCookieValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithCookieValues(names...))
		return nil
	}
}

// WithPathValues sets the path values for the default decoder.
// This option is only applicable when using the default decoder.
func WithPathValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithPathValues(names...))
		return nil
	}
}

// WithSessionValues is a convenience option that adds [reflectd.WithSessionValues] to the decoder options.
// This option is only applicable when using the default decoder.
func WithSessionValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, reflectd.WithSessionValues(names...))
		return nil
	}
}

// WithMiddleware adds [http.Handler] middleware wrappers
// to the adaptor. They will be applied in reverse order,
// so the first middleware is the outermost wrapper.
func WithMiddleware(m ...Middleware) Option {
	return func(o *options) error {
		for _, mw := range m {
			if mw == nil {
				return fmt.Errorf("nil adaptor middleware")
			}
		}
		o.Middleware = append(o.Middleware, m...)
		return nil
	}
}

// WithMiddlewareIncludingOnly sets the middleware for the adaptor,
// replacing any existing middleware with the provided ones.
//
// They will be applied in reverse order,
// so the first middleware is the outermost wrapper.
func WithMiddlewareIncludingOnly(m ...Middleware) Option {
	return func(o *options) error {
		for _, mw := range m {
			if mw == nil {
				return fmt.Errorf("nil adaptor middleware")
			}
		}
		o.Middleware = m
		return nil
	}
}
