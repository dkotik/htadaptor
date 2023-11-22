package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func NewNullaryFuncAdaptor[O any](
	domainCall func(context.Context) (O, error),
	encoder Encoder[O],
) (*NullaryFuncAdaptor[O], error) {
	if domainCall == nil {
		return nil, errors.New("cannot use a <nil> domain call")
	}
	var zero Encoder[O]
	if encoder == zero {
		return nil, errors.New("cannot use a <nil> encoder")
	}
	return &NullaryFuncAdaptor[O]{
		domainCall: domainCall,
		encoder:    encoder,
	}, nil
}

type NullaryFuncAdaptor[O any] struct {
	domainCall func(context.Context) (O, error)
	encoder    Encoder[O]
}

func (a *NullaryFuncAdaptor[O]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	response, err := a.domainCall(r.Context())
	if err != nil {
		return err
	}
	if err = a.encoder.Encode(w, response); err != nil {
		return fmt.Errorf("unable to encode: %w", err)
	}
	return nil
}
