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
			if o.ResponseHandler == nil {
				if err = WithDefaultResponseHandler()(o); err != nil {
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
		domainCall:      domainCall,
		encoder:         o.Encoder,
		responseHandler: o.ResponseHandler,
	}, nil
}

type NullaryFuncAdaptor[O any] struct {
	domainCall      func(context.Context) (O, error)
	encoder         Encoder
	responseHandler ResponseHandler
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
		return NewEncodingError(err)
	}
	return a.responseHandler.HandleSuccess(w, r)
}

func (a *NullaryFuncAdaptor[O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	if err != nil {
		a.responseHandler.HandleError(w, r, err)
	}
}
