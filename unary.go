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
	decoder Decoder[T, V, O],
) (*UnaryFuncAdaptor[T, V, O], error) {
	if domainCall == nil {
		return nil, errors.New("cannot use a <nil> domain call")
	}
	var zero Decoder[T, V, O]
	if decoder == zero {
		return nil, errors.New("cannot use a <nil> decoder")
	}
	return &UnaryFuncAdaptor[T, V, O]{
		domainCall: domainCall,
		decoder:    decoder,
	}, nil
}

type UnaryFuncAdaptor[
	T any,
	V Validatable[T],
	O any,
] struct {
	domainCall func(context.Context, V) (O, error)
	decoder    Decoder[T, V, O]
}

func (a *UnaryFuncAdaptor[T, V, O]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	request, encoder, err := a.decoder.Decode(w, r)
	if err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to decode: %w", err))
	}
	if err = request.Validate(); err != nil {
		return NewInvalidRequestError(err)
	}

	response, err := a.domainCall(r.Context(), request)
	if err != nil {
		return err
	}
	if err = encoder.Encode(w, response); err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	return nil
}

type StringUnaryFuncAdaptor[O any] struct {
	domainCall func(context.Context, string) (O, error)
	extractor  func(*http.Request) (string, error)
	encoder    Encoder[O]
}

func NewStringUnaryFuncAdaptor[O any](
	domainCall func(context.Context, string) (O, error),
	extractor func(*http.Request) (string, error),
	encoder Encoder[O],
) (*StringUnaryFuncAdaptor[O], error) {
	if domainCall == nil {
		return nil, errors.New("cannot use a <nil> domain call")
	}
	if extractor == nil {
		return nil, errors.New("cannot use a <nil> string extractor")
	}
	var zero Encoder[O]
	if encoder == zero {
		return nil, errors.New("cannot use a <nil> encoder")
	}
	return &StringUnaryFuncAdaptor[O]{
		domainCall: domainCall,
		extractor:  extractor,
		encoder:    encoder,
	}, nil
}

func (a *StringUnaryFuncAdaptor[O]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	request, err := a.extractor(r)
	if err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to extract string: %w", err))
	}

	response, err := a.domainCall(r.Context(), request)
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, response); err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	return nil
}
