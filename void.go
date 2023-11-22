package htadaptor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

func NewVoidFuncAdaptor[
	T any,
	V Validatable[T],
](
	domainCall func(context.Context, V) error,
	withOptions ...Option,
) (*VoidFuncAdaptor[T, V], error) {
	o := &options{}
	err := WithOptions(append(
		withOptions,
		func(o *options) (err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("unable to apply default option: %w", err)
				}
			}()

			if o.Decoder == nil {
				if err = WithDefaultDecoder()(o); err != nil {
					return err
				}
			}
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
			return nil
		},
	)...)(o)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize void adaptor: %w", err)
	}

	return &VoidFuncAdaptor[T, V]{
		domainCall:   domainCall,
		decoder:      o.Decoder,
		errorHandler: o.ErrorHandler,
		logger:       o.Logger,
	}, nil
}

type VoidFuncAdaptor[
	T any,
	V Validatable[T],
] struct {
	domainCall   func(context.Context, V) error
	decoder      Decoder
	errorHandler ErrorHandler
	logger       *slog.Logger
}

func (a *VoidFuncAdaptor[T, V]) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	var request V
	if err := a.decoder.Decode(request, r); err != nil {
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

func (a *VoidFuncAdaptor[T, V]) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := a.ServeHyperText(w, r)
	if err == nil {
		a.logger.ErrorContext(
			r.Context(),
			err.Error(),
			slog.String("method", r.Method),
			slog.String("URL", r.URL.String()),
		)
	} else {
		a.logger.DebugContext(
			r.Context(),
			"completed HTTP request via void adaptor",
			slog.String("method", r.Method),
			slog.String("URL", r.URL.String()),
		)
	}
}

/*
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
*/
