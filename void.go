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
		domainCall:   domainCall,
		decoder:      o.Decoder,
		errorHandler: o.ErrorHandler,
		logger:       o.Logger,
	}, nil
}

type VoidFuncAdaptor[
	T any,
	V Validatable[T],
] struct {
	domainCall   func(context.Context, V) error
	decoder      Decoder
	errorHandler ErrorHandler
	logger       Logger
}

func (a *VoidFuncAdaptor[T, V]) executeDomainCall(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	var request V = new(T)
	if err := a.decoder.Decode(request, r); err != nil {
		return NewDecodingError(err)
	}
	if err = request.Validate(); err != nil {
		return NewInvalidRequestError(err)
	}
	if err = a.domainCall(r.Context(), request); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (a *VoidFuncAdaptor[T, V]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.executeDomainCall(w, r)
	if err != nil {
		err = a.errorHandler.HandleError(w, r, err)
	}
	a.logger.LogRequest(r, err)
}
