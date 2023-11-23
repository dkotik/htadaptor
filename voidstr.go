package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func NewVoidStringFuncAdaptor(
	domainCall func(context.Context, string) error,
	stringExtractor StringExtractor,
	withOptions ...Option,
) (*VoidStringFuncAdaptor, error) {
	o := &options{}
	err := WithOptions(append(
		withOptions,
		func(o *options) (err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("unable to apply default option: %w", err)
				}
			}()

			if o.ErrorHandler == nil {
				if err = WithDefaultErrorHandler()(o); err != nil {
					return err
				}
			}
			if o.Logger == nil {
				if err = WithDefaultLogger()(o); err != nil {
					return err
				}
			}
			if domainCall == nil {
				return errors.New("cannot use a <nil> domain call")
			}
			if stringExtractor == nil {
				return errors.New("cannot use a <nil> string extractor")
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

type VoidStringFuncAdaptor struct {
	domainCall      func(context.Context, string) error
	stringExtractor StringExtractor
	errorHandler    ErrorHandler
	logger          RequestLogger
}

func (a *VoidStringFuncAdaptor) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	value, err := a.stringExtractor(r)
	if err != nil {
		return NewInvalidRequestError(fmt.Errorf("unable to decode: %w", err))
	}
	if err = a.domainCall(r.Context(), value); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (a *VoidStringFuncAdaptor) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	a.logger.LogRequest(r, err)
	if err != nil {
		a.errorHandler.HandleError(w, r, err)
	}
}
