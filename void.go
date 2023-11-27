package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func NewVoidFuncAdaptor[
	T any,
	V Validatable[T],
](
	domainCall func(context.Context, V) error,
	withOptions ...Option,
) (*VoidFuncAdaptor[T, V], error) {
	o := &options{}
	err := WithOptions(append(
		withOptions,
		func(o *options) (err error) {
			if err = o.Validate(); err != nil {
				return err
			}
			if o.Decoder == nil {
				if err = WithDefaultDecoder()(o); err != nil {
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
		return nil, fmt.Errorf("unable to initialize void adaptor: %w", err)
	}

	return &VoidFuncAdaptor[T, V]{
		domainCall:      domainCall,
		decoder:         o.Decoder,
		responseHandler: o.ResponseHandler,
	}, nil
}

type VoidFuncAdaptor[
	T any,
	V Validatable[T],
] struct {
	domainCall      func(context.Context, V) error
	decoder         Decoder
	responseHandler ResponseHandler
}

func (a *VoidFuncAdaptor[T, V]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	var request V = new(T)
	if err := a.decoder.Decode(request, r); err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to decode: %w", err))
	}
	if err = request.Validate(); err != nil {
		return NewInvalidRequestError(err)
	}
	if err = a.domainCall(r.Context(), request); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return a.responseHandler.HandleSuccess(w, r)
}

func (a *VoidFuncAdaptor[T, V]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	if err != nil {
		a.responseHandler.HandleError(w, r, err)
	}
}
