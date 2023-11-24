package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func NewNullaryFuncAdaptor[O any](
	domainCall func(context.Context) (O, error),
	withOptions ...Option,
) (*NullaryFuncAdaptor[O], error) {
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
			if domainCall == nil {
				return errors.New("cannot use a <nil> domain call")
			}
			return nil
		},
	)...)(o)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize nullary adaptor: %w", err)
	}

	return &NullaryFuncAdaptor[O]{
		domainCall:   domainCall,
		encoder:      o.Encoder,
		errorHandler: o.ErrorHandler,
		logger:       o.Logger,
	}, nil
}

type NullaryFuncAdaptor[O any] struct {
	domainCall   func(context.Context) (O, error)
	encoder      Encoder
	errorHandler ErrorHandler
	logger       RequestLogger
}

func (a *NullaryFuncAdaptor[O]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	response, err := a.domainCall(r.Context())
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, response); err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	return nil
}

func (a *NullaryFuncAdaptor[O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	a.logger.LogRequest(r, err)
	if err != nil {
		a.errorHandler.HandleError(w, r, err)
	}
}
