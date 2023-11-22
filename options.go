package htadaptor

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	"log/slog"

	"github.com/dkotik/htadaptor/decoder"
)

type options struct {
	Decoder      Decoder
	Encoder      Encoder
	ErrorHandler ErrorHandler
	Logger       *slog.Logger
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

func WithDecoderOptions(withOptions ...decoder.Option) Option {
	return func(o *options) error {
		d, err := decoder.New(withOptions...)
		if err != nil {
			return err
		}
		return WithDecoder(d)(o)
	}
}

func WithDefaultDecoder() Option {
	return WithDecoderOptions()
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

func WithDefaultErrorHandler() Option {
	return WithErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		htError, ok := err.(Error)
		if ok {
			http.Error(w, err.Error(), htError.HyperTextStatusCode())
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func WithLogger(l *slog.Logger) Option {
	return func(o *options) error {
		if l == nil {
			return errors.New("cannot use a <nil> logger")
		}
		if o.Logger != nil {
			return errors.New("logger is already set")
		}
		o.Logger = l
		return nil
	}
}

func WithDefaultLogger() Option {
	return WithLogger(slog.Default())
}
