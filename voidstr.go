package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dkotik/htadaptor/extract"
)

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
			if o.ResponseHandler == nil {
				if err = WithDefaultResponseHandler()(o); err != nil {
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
		return nil, fmt.Errorf("unable to initialize void adaptor: %w", err)
	}

	return &VoidStringFuncAdaptor{
		domainCall:      domainCall,
		stringExtractor: stringExtractor,
		responseHandler: o.ResponseHandler,
	}, nil
}

type VoidStringFuncAdaptor struct {
	domainCall      func(context.Context, string) error
	stringExtractor extract.StringValueExtractor
	responseHandler ResponseHandler
}

func (a *VoidStringFuncAdaptor) ServeHyperText(
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
	return a.responseHandler.HandleSuccess(w, r)
}

func (a *VoidStringFuncAdaptor) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	if err != nil {
		a.responseHandler.HandleError(w, r, err)
	}
}
