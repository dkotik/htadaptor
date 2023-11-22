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
	decoder Decoder[T, V, T],
) (*VoidFuncAdaptor[T, V], error) {
	if domainCall == nil {
		return nil, errors.New("cannot use a <nil> domain call")
	}
	var zero Decoder[T, V, T]
	if decoder == zero {
		return nil, errors.New("cannot use a <nil> decoder")
	}
	return &VoidFuncAdaptor[T, V]{
		domainCall: domainCall,
		decoder:    decoder,
	}, nil
}

type VoidFuncAdaptor[
	T any,
	V Validatable[T],
] struct {
	domainCall func(context.Context, V) error
	decoder    Decoder[T, V, T]
}

func (a *VoidFuncAdaptor[T, V]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	request, _, err := a.decoder.Decode(w, r)
	if err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to decode: %w", err))
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

type StringVoidFuncAdaptor struct {
	domainCall func(context.Context, string) error
	extractor  func(*http.Request) (string, error)
}

func NewStringVoidFuncAdaptor(
	domainCall func(context.Context, string) error,
	extractor func(*http.Request) (string, error),
) (*StringVoidFuncAdaptor, error) {
	if domainCall == nil {
		return nil, errors.New("cannot use a <nil> domain call")
	}
	if extractor == nil {
		return nil, errors.New("cannot use a <nil> string extractor")
	}
	return &StringVoidFuncAdaptor{
		domainCall: domainCall,
		extractor:  extractor,
	}, nil
}

func (a *StringVoidFuncAdaptor) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	request, err := a.extractor(r)
	if err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to extract string: %w", err))
	}
	if err = a.domainCall(r.Context(), request); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
