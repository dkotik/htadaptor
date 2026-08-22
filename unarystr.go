package htadaptor

import (
	"context"
	"errors"
	"net/http"

	"github.com/dkotik/htadaptor/extract"
)

// AdaptStringFunc creates a new adaptor for a
// function that takes a string and returns a struct.
func (a Adaptor) AdaptStringFunc[O any](
	domainCall func(context.Context, string) (O, error),
	stringExtractor extract.StringValueExtractor,
	withOptions ...Option,
) (http.Handler, error) {
	if domainCall == nil {
		return nil, errors.New("nil domain call")
	}
	if stringExtractor == nil {
		return nil, errors.New("nil string extractor")
	}
	o, err := a.initialize(withOptions)
	if err != nil {
		return nil, err
	}
	return &UnaryStringFuncAdaptor[O]{
		domainCall:      domainCall,
		stringExtractor: stringExtractor,
		statusCode:      o.StatusCode,
		encoder:         o.Encoder,
		errorHandler:    o.ErrorHandler,
	}, nil
}

// UnaryStringFuncAdaptor extracts a string value from request
// and calls a domain function with it expecting
// a struct response.
type UnaryStringFuncAdaptor[O any] struct {
	domainCall      func(context.Context, string) (O, error)
	stringExtractor extract.StringValueExtractor
	statusCode      int
	encoder         Encoder
	errorHandler    ErrorHandler
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
	if err = a.encoder.Encode(w, r, a.statusCode, response); err != nil {
		return NewEncodingError(err)
	}
	return nil
}

// ServeHTTP satisfies [http.Handler] interface.
func (a *UnaryStringFuncAdaptor[O]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.executeDomainCall(w, r)
	if err != nil {
		err = a.errorHandler.HandleError(w, r, err)
	}
}
