package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dkotik/htadaptor/extract"
)

var ErrNoStringValue = errors.New("empty value")

func NewUnaryStringFuncAdaptor[O any](
	domainCall func(context.Context, string) (O, error),
	stringExtractor extract.StringValueExtractor,
	withOptions ...Option,
) (*UnaryStringFuncAdaptor[O], error) {
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
			if domainCall == nil {
				return errors.New("cannot use a <nil> domain call")
			}
			if stringExtractor == nil {
				return errors.New("cannot use a <nil> string value extractor")
			}
			return nil
		},
	)...)(o)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize unary adaptor: %w", err)
	}

	return &UnaryStringFuncAdaptor[O]{
		domainCall:      domainCall,
		stringExtractor: stringExtractor,
		encoder:         o.Encoder,
		errorHandler:    o.ErrorHandler,
		logger:          o.Logger,
	}, nil
}

type UnaryStringFuncAdaptor[O any] struct {
	domainCall      func(context.Context, string) (O, error)
	stringExtractor extract.StringValueExtractor
	encoder         Encoder
	errorHandler    ErrorHandler
	logger          Logger
}

func (a *UnaryStringFuncAdaptor[O]) executeDomainCall(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	value, err := a.stringExtractor.ExtractStringValue(r)
	if err != nil {
		return NewDecodingError(err)
	}
	response, err := a.domainCall(r.Context(), value)
	if err != nil {
		return err
	}
	writeEncoderContentType(w, a.encoder)
	if err = a.encoder.Encode(w, response); err != nil {
		return NewEncodingError(err)
	}
	return nil
}

func (a *UnaryStringFuncAdaptor[O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.executeDomainCall(w, r)
	if err != nil {
		err = a.errorHandler.HandleError(w, r, err)
	}
	a.logger.LogRequest(r, err)
}
