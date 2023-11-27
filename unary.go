package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func NewUnaryFuncAdaptor[
	T any,
	V Validatable[T],
	O any,
](
	domainCall func(context.Context, V) (O, error),
	withOptions ...Option,
) (*UnaryFuncAdaptor[T, V, O], error) {
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
		return nil, fmt.Errorf("unable to initialize unary adaptor: %w", err)
	}

	return &UnaryFuncAdaptor[T, V, O]{
		domainCall:      domainCall,
		decoder:         o.Decoder,
		encoder:         o.Encoder,
		responseHandler: o.ResponseHandler,
	}, nil
}

type UnaryFuncAdaptor[
	T any,
	V Validatable[T],
	O any,
] struct {
	domainCall      func(context.Context, V) (O, error)
	decoder         Decoder
	encoder         Encoder
	responseHandler ResponseHandler
}

func (a *UnaryFuncAdaptor[T, V, O]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	var request V = new(T)
	// request := new(V)
	if err = a.decoder.Decode(request, r); err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to decode: %w", err))
	}
	if err = request.Validate(); err != nil {
		return NewInvalidRequestError(err)
	}

	response, err := a.domainCall(r.Context(), request)
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, response); err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	return a.responseHandler.HandleSuccess(w, r)
}

func (a *UnaryFuncAdaptor[T, V, O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	if err != nil {
		a.responseHandler.HandleError(w, r, err)
	}
}
