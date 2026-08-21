package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// AdaptNullaryFunc creates a new adaptor for a
// function that takes no input and returns a struct.
func (a *Adaptor) AdaptNullaryFunc[O any](
	domainCall func(context.Context) (O, error),
) http.Handler {
	a.panicIfUninitialized()
	if domainCall == nil {
		panic("nil domain call")
	}
	return &NullaryFuncAdaptor[O]{
		domainCall:   domainCall,
		statusCode:   a.statusCode,
		encoder:      a.encoder,
		errorHandler: a.errorHandler,
	}
}

// NewNullaryFuncAdaptor creates a new adaptor for a
// function that takes no input and returns a struct.
//
// DEPRECATED: Use [Adaptor.AdaptNullaryFunc] instead.
func NewNullaryFuncAdaptor[O any](
	domainCall func(context.Context) (O, error),
	withOptions ...Option,
) (*NullaryFuncAdaptor[O], error) {
	o := &options{}
	err := WithOptions(append(
		withOptions,
		WithDefaultStatusCode(),
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
		statusCode:   o.StatusCode,
		encoder:      o.Encoder,
		errorHandler: o.ErrorHandler,
	}, nil
}

// NullaryFuncAdaptor calls a domain function with no input
// and returns a response struct.
type NullaryFuncAdaptor[O any] struct {
	domainCall   func(context.Context) (O, error)
	statusCode   int
	encoder      Encoder
	errorHandler ErrorHandler
}

func (a *NullaryFuncAdaptor[O]) executeDomainCall(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	response, err := a.domainCall(r.Context())
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, r, a.statusCode, response); err != nil {
		return NewEncodingError(err)
	}
	return nil
}

// ServeHTTP satisfies [http.Handler] interface.
func (a *NullaryFuncAdaptor[O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.executeDomainCall(w, r)
	if err != nil {
		err = a.errorHandler.HandleError(w, r, err)
	}
}
