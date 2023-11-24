package htadaptor

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"log/slog"
)

type options struct {
	Decoder        Decoder
	DecoderOptions []StructDecoderOption
	Encoder        Encoder
	ErrorHandler   ErrorHandler
	Logger         RequestLogger
}

func (o *options) Validate() (err error) {
	if len(o.DecoderOptions) > 0 {
		if o.Decoder != nil {
			return fmt.Errorf("option WithDecoder conflicts with %d decoder options; provide either a prepared decoder or options for preparing one, but not both", len(o.DecoderOptions))
		}
		o.Decoder, err = NewStructDecoder(o.DecoderOptions...)
		if err != nil {
			return err
		}
	}
	if o.ErrorHandler == nil {
		if err = WithDefaultErrorHandler()(o); err != nil {
			return err
		}
	}
	if o.Logger == nil {
		if err = WithDefaultLogger()(o); err != nil {
			return err
		}
	}
	return nil
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
			return errors.New("cannot use a <nil> encoder")
		}
		if o.Encoder != nil {
			return errors.New("encoder is already set")
		}
		o.Encoder = e
		return nil
	}
}

func WithEncoderFunc(f func(http.ResponseWriter, any) error) Option {
	return WithEncoder(EncoderFunc(f))
}

func WithHyperTextEncoder(t *template.Template) Option {
	return func(o *options) error {
		if t == nil {
			return errors.New("cannot use a <nil> template")
		}
		return WithEncoderFunc(func(w http.ResponseWriter, v any) error {
			w.Header().Set("Content-Type", "text/html")
			return t.Execute(w, v)
		})(o)
	}
}

func WithDefaultEncoder() Option {
	return WithEncoderFunc(func(w http.ResponseWriter, v any) error {
		// panic("encoder")
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(v)
	})
}

func WithErrorHandler(e ErrorHandler) Option {
	return func(o *options) error {
		if e == nil {
			return errors.New("cannot use a <nil> error handler")
		}
		if o.ErrorHandler != nil {
			return errors.New("error handler is already set")
		}
		o.ErrorHandler = e
		return nil
	}
}

func WithErrorHandlerFunc(f func(http.ResponseWriter, *http.Request, error)) Option {
	return WithErrorHandler(ErrorHandlerFunc(f))
}

var defaultErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	htError, ok := err.(Error)
	if ok {
		http.Error(w, err.Error(), htError.HyperTextStatusCode())
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WithDefaultErrorHandler() Option {
	return WithErrorHandlerFunc(defaultErrorHandler)
}

func WithLogger(l RequestLogger) Option {
	return func(o *options) error {
		if l == nil {
			return errors.New("cannot use a <nil> request logger")
		}
		if o.Logger != nil {
			return errors.New("request logger is already set")
		}
		o.Logger = l
		return nil
	}
}

func WithSlogLogger(l *slog.Logger, successLevel slog.Leveler) Option {
	return func(o *options) error {
		if l == nil {
			return errors.New("cannot use a <nil> structured logger")
		}
		return WithLogger(NewRequestLogger(l, successLevel))(o)
	}
}

func WithDefaultLogger() Option {
	return WithLogger(NewRequestLogger(slog.Default(), slog.LevelInfo))
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

func WithDecoderOptions(withOptions ...StructDecoderOption) Option {
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
		d, err := NewStructDecoder()
		if err != nil {
			return err
		}
		return WithDecoder(d)(o)
	}
}

func WithReadLimit(upto int64) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, WithDecoderReadLimit(upto))
		return nil
	}
}

func WithMemoryLimit(upto int64) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, WithDecoderMemoryLimit(upto))
		return nil
	}
}

func WithExtractors(exs ...RequestValueExtractor) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, WithDecoderExtractors(exs...))
		return nil
	}
}

func WithQueryValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, WithDecoderQueryValues(names...))
		return nil
	}
}

func WithHeaderValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, WithDecoderHeaderValues(names...))
		return nil
	}
}

func WithPathValues(names ...string) Option {
	return func(o *options) error {
		o.DecoderOptions = append(o.DecoderOptions, WithDecoderPathValues(names...))
		return nil
	}
}
