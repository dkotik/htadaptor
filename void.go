package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// NewVoidFuncAdaptor creates a new adaptor for a
// function that takes a decoded request and returns nothing.
func NewVoidFuncAdaptor[T any, V *T](
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

// VoidStringFuncAdaptor calls a domain function with decoded
// request without returning no response other than an error.
type VoidFuncAdaptor[T any, V *T] struct {
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

	ctx := r.Context()
	if validatable, ok := any(request).(Validatable); ok {
		if err = validatable.Validate(ctx); err != nil {
			return err
		}
	}
	if err = a.domainCall(ctx, request); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// ServeHTTP satisfies [http.Handler] interface.
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
