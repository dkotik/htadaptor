package htadaptor

import (
	"context"
	"errors"
	"net/http"
)

// AdaptVoidFunc creates a new adaptor for a
// function that takes a decoded request and returns nothing.
func (a Adaptor) AdaptVoidFunc[T any, V Validatable[T]](
	domainCall func(context.Context, V) error,
	withOptions ...Option,
) (http.Handler, error) {
	if domainCall == nil {
		return nil, errors.New("nil domain call")
	}
	o, err := a.initialize(withOptions)
	if err != nil {
		return nil, err
	}
	return &VoidFuncAdaptor[T, V]{
		domainCall: domainCall,
		// statusCode:   a.statusCode,
		// encoder:      a.encoder,
		decoder:      o.Decoder,
		errorHandler: o.ErrorHandler,
	}, nil
}

// VoidStringFuncAdaptor calls a domain function with decoded
// request without returning no response other than an error.
type VoidFuncAdaptor[T any, V Validatable[T]] struct {
	domainCall   func(context.Context, V) error
	decoder      Decoder
	errorHandler ErrorHandler
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
	if err = request.Validate(ctx); err != nil {
		return err
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
}
