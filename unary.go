package htadaptor

import (
	"context"
	"errors"
	"net/http"
)

// AdaptFunc creates a new adaptor for a
// function that takes a validatable struct and returns a struct.
func (a Adaptor) AdaptFunc[T any, V *T, O any](
	domainCall func(context.Context, V) (O, error),
	withOptions ...Option,
) (http.Handler, error) {
	if domainCall == nil {
		return nil, errors.New("nil domain call")
	}
	o, err := a.initialize(withOptions)
	if err != nil {
		return nil, err
	}
	return &UnaryFuncAdaptor[T, V, O]{
		domainCall:   domainCall,
		statusCode:   o.StatusCode,
		encoder:      o.Encoder,
		decoder:      o.Decoder,
		errorHandler: o.ErrorHandler,
	}, nil
}

// UnaryFuncAdaptor extracts a struct from request
// and calls a domain function with it expecting
// a struct response.
type UnaryFuncAdaptor[T any, V *T, O any] struct {
	domainCall   func(context.Context, V) (O, error)
	statusCode   int
	decoder      Decoder
	encoder      Encoder
	errorHandler ErrorHandler
}

func (a *UnaryFuncAdaptor[T, V, O]) executeDomainCall(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	var request V = new(T)
	if err = a.decoder.Decode(request, r); err != nil {
		return NewDecodingError(err)
	}

	ctx := r.Context()
	if validatable, ok := any(request).(Validatable); ok {
		if err = validatable.Validate(ctx); err != nil {
			return err
		}
	}
	response, err := a.domainCall(ctx, request)
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, r, a.statusCode, response); err != nil {
		return NewEncodingError(err)
	}
	return nil
}

// ServeHTTP satisfies [http.Handler] interface.
func (a *UnaryFuncAdaptor[T, V, O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.executeDomainCall(w, r)
	if err != nil {
		err = a.errorHandler.HandleError(w, r, err)
	}
}
