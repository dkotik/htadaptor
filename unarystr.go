package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func NewUnaryStringFuncAdaptor[O any](
	domainCall func(context.Context, string) (O, error),
	stringExtractor StringValueExtractor,
	withOptions ...Option,
) (*UnaryStringFuncAdaptor[O], error) {
	o := &options{}
	err := WithOptions(append(
		withOptions,
		func(o *options) (err error) {
			if err = o.Validate(); err != nil {
				return err
			}

			if o.Encoder == nil {
				if err = WithDefaultEncoder()(o); err != nil {
					return err
				}
			}
			if o.ResponseHandler == nil {
				if err = WithDefaultResponseHandler()(o); err != nil {
					return err
				}
			}
			if domainCall == nil {
				return errors.New("cannot use a <nil> domain call")
			}
			if stringExtractor == nil {
				return errors.New("cannot use a <nil> string value extractor")
			}
			return nil
		},
	)...)(o)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize unary adaptor: %w", err)
	}

	return &UnaryStringFuncAdaptor[O]{
		domainCall:      domainCall,
		stringExtractor: stringExtractor,
		encoder:         o.Encoder,
		responseHandler: o.ResponseHandler,
	}, nil
}

type UnaryStringFuncAdaptor[O any] struct {
	domainCall      func(context.Context, string) (O, error)
	stringExtractor StringValueExtractor
	encoder         Encoder
	responseHandler ResponseHandler
}

func (a *UnaryStringFuncAdaptor[O]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	value, err := a.stringExtractor.ExtractStringValue(r)
	if err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to decode string value: %w", err))
	}
	response, err := a.domainCall(r.Context(), value)
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, response); err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	return a.responseHandler.HandleSuccess(w, r)
}

func (a *UnaryStringFuncAdaptor[O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	if err != nil {
		a.responseHandler.HandleError(w, r, err)
	}
}
