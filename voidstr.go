package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dkotik/htadaptor/extract"
)

// NewVoidStringFuncAdaptor creates a new adaptor for a
// function that takes a string and returns nothing.
func NewVoidStringFuncAdaptor(
	domainCall func(context.Context, string) error,
	stringExtractor extract.StringValueExtractor,
	withOptions ...Option,
) (*VoidStringFuncAdaptor, error) {
	o := &options{}
	err := WithOptions(append(
		withOptions,
		func(o *options) (err error) {
			if err = o.Validate(); err != nil {
				return err
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
		return nil, fmt.Errorf("unable to initialize void adaptor: %w", err)
	}

	return &VoidStringFuncAdaptor{
		domainCall:      domainCall,
		stringExtractor: stringExtractor,
		errorHandler:    o.ErrorHandler,
		logger:          o.Logger,
	}, nil
}

// VoidStringFuncAdaptor extracts a string value from request
// and calls a domain function with it without expecting no response
// other than an error value.
type VoidStringFuncAdaptor struct {
	domainCall      func(context.Context, string) error
	stringExtractor extract.StringValueExtractor
	errorHandler    ErrorHandler
	logger          Logger
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
	a.logger.LogRequest(r, err)
}
