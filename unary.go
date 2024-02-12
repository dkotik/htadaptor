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
		domainCall:   domainCall,
		decoder:      o.Decoder,
		encoder:      o.Encoder,
		errorHandler: o.ErrorHandler,
		logger:       o.Logger,
	}, nil
}

type UnaryFuncAdaptor[
	T any,
	V Validatable[T],
	O any,
] struct {
	domainCall   func(context.Context, V) (O, error)
	decoder      Decoder
	encoder      Encoder
	errorHandler ErrorHandler
	logger       Logger
}

func (a *UnaryFuncAdaptor[T, V, O]) executeDomainCall(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	var request V = new(T)
	// request := new(V)
	if err = a.decoder.Decode(request, r); err != nil {
		return NewDecodingError(err)
	}
	if err = request.Validate(); err != nil {
		return NewInvalidRequestError(err)
	}

	response, err := a.domainCall(r.Context(), request)
	if err != nil {
		return err
	}
	writeEncoderContentType(w, a.encoder)
	if err = a.encoder.Encode(w, response); err != nil {
		return NewEncodingError(err)
	}
	return nil
}

func (a *UnaryFuncAdaptor[T, V, O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.executeDomainCall(w, r)
	if err != nil {
		err = a.errorHandler.HandleError(w, r, err)
	}
	a.logger.LogRequest(r, err)
}
