package htadaptor

import (
	"context"
	"errors"
	"net/http"
)

// AdaptNullaryFunc creates a new adaptor for a
// function that takes no input and returns a struct.
func (a Adaptor) AdaptNullaryFunc[O any](
	domainCall func(context.Context) (O, error),
	withOptions ...Option,
) (http.Handler, error) {
	if domainCall == nil {
		return nil, errors.New("nil domain call")
	}
	o, err := a.initialize(withOptions)
	if err != nil {
		return nil, err
	}
	return ApplyMiddleware(
		&NullaryFuncAdaptor[O]{
			domainCall:   domainCall,
			statusCode:   o.StatusCode,
			encoder:      o.Encoder,
			errorHandler: o.ErrorHandler,
		}, o.Middleware...), nil
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
