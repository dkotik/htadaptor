package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

func NewUnaryFuncAdaptor[
	T any,
	V Validatable[T],
	O any,
](
	domainCall func(context.Context, V) (O, error),
	withOptions ...Option,
) (*UnaryFuncAdaptor[T, V, O], error) {
	o := &options{}
	err := WithOptions(append(
		withOptions,
		func(o *options) (err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("unable to apply default option: %w", err)
				}
			}()

			if o.Encoder == nil {
				if err = WithDefaultEncoder()(o); err != nil {
					return err
				}
			}
			if o.Decoder == nil {
				if err = WithDefaultDecoder()(o); err != nil {
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
			if domainCall == nil {
				return errors.New("cannot use a <nil> domain call")
			}
			return nil
		},
	)...)(o)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize unary adaptor: %w", err)
	}

	return &UnaryFuncAdaptor[T, V, O]{
		domainCall:   domainCall,
		decoder:      o.Decoder,
		encoder:      o.Encoder,
		errorHandler: o.ErrorHandler,
		logger:       o.Logger,
	}, nil
}

type UnaryFuncAdaptor[
	T any,
	V Validatable[T],
	O any,
] struct {
	domainCall   func(context.Context, V) (O, error)
	decoder      Decoder
	encoder      Encoder
	errorHandler ErrorHandler
	logger       *slog.Logger
}

func (a *UnaryFuncAdaptor[T, V, O]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	var request V
	if err = a.decoder.Decode(request, r); err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to decode: %w", err))
	}
	if err = request.Validate(); err != nil {
		return NewInvalidRequestError(err)
	}

	response, err := a.domainCall(r.Context(), request)
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, response); err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	return nil
}

func (a *UnaryFuncAdaptor[T, V, O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	if err == nil {
		a.logger.ErrorContext(
			r.Context(),
			err.Error(),
			slog.String("method", r.Method),
			slog.String("URL", r.URL.String()),
		)
	} else {
		a.logger.DebugContext(
			r.Context(),
			"completed HTTP request via unary adaptor",
			slog.String("method", r.Method),
			slog.String("URL", r.URL.String()),
		)
	}
}

/*
type StringUnaryFuncAdaptor[O any] struct {
	domainCall func(context.Context, string) (O, error)
	extractor  func(*http.Request) (string, error)
	encoder    Encoder[O]
}

func NewStringUnaryFuncAdaptor[O any](
	domainCall func(context.Context, string) (O, error),
	extractor func(*http.Request) (string, error),
	encoder Encoder[O],
) (*StringUnaryFuncAdaptor[O], error) {
	if domainCall == nil {
		return nil, errors.New("cannot use a <nil> domain call")
	}
	if extractor == nil {
		return nil, errors.New("cannot use a <nil> string extractor")
	}
	var zero Encoder[O]
	if encoder == zero {
		return nil, errors.New("cannot use a <nil> encoder")
	}
	return &StringUnaryFuncAdaptor[O]{
		domainCall: domainCall,
		extractor:  extractor,
		encoder:    encoder,
	}, nil
}

func (a *StringUnaryFuncAdaptor[O]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	request, err := a.extractor(r)
	if err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to extract string: %w", err))
	}

	response, err := a.domainCall(r.Context(), request)
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, response); err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	return nil
}
*/
