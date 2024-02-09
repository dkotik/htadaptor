package htadaptor

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dkotik/htadaptor/extract"
	"github.com/dkotik/htadaptor/reflectd"
	"github.com/dkotik/htadaptor/slogrh"
)

type options struct {
	Decoder                Decoder
	DecoderOptions         []reflectd.Option
	Encoder                Encoder
	ResponseHandler        ResponseHandler
	ResponseHandlerOptions []slogrh.Option
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

	if len(o.ResponseHandlerOptions) > 0 {
		if o.ResponseHandler != nil {
			return fmt.Errorf("option WithResponseHandler conflicts with %d response handler options; provide either a prepared response handler or options for preparing an slog one, but not both", len(o.ResponseHandlerOptions))
		}
		o.ResponseHandler, err = slogrh.New(o.ResponseHandlerOptions...)
		if err != nil {
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

func WithResponseHandler(h ResponseHandler) Option {
	return func(o *options) error {
		if h == nil {
			return errors.New("cannot use a <nil> response handler")
		}
		if o.ResponseHandler != nil {
			return errors.New("response handler is already set")
		}
		o.ResponseHandler = h
		return nil
	}
}

func WithDefaultResponseHandler() Option {
	return func(o *options) (err error) {
		rh, err := slogrh.New()
		if err != nil {
			return err
		}
		return WithResponseHandler(rh)(o)
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
