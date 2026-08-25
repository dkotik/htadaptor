package htadaptor

import (
	"context"
	"errors"
	"net/http"

	"github.com/dkotik/htadaptor/extract"
)

// AdaptStringFunc creates a new adaptor for a
// function that takes a string and returns nothing.
func (a Adaptor) AdaptVoidStringFunc(
	domainCall func(context.Context, string) error,
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
	return ApplyMiddleware(
		&VoidStringFuncAdaptor{
			domainCall:      domainCall,
			stringExtractor: stringExtractor,
			errorHandler:    o.ErrorHandler,
		}, o.Middleware...), nil
}

// VoidStringFuncAdaptor extracts a string value from request
// and calls a domain function with it without expecting no response
// other than an error value.
type VoidStringFuncAdaptor struct {
	domainCall      func(context.Context, string) error
	stringExtractor extract.StringValueExtractor
	errorHandler    ErrorHandler
}

func (a *VoidStringFuncAdaptor) executeDomainCall(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	value, err := a.stringExtractor.ExtractStringValue(r)
	if err != nil {
		return NewDecodingError(err)
	}
	if err = a.domainCall(r.Context(), value); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// ServeHTTP satisfies [http.Handler] interface.
func (a *VoidStringFuncAdaptor) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.executeDomainCall(w, r)
	if err != nil {
		err = a.errorHandler.HandleError(w, r, err)
	}
}
