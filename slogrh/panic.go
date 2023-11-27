package slogrh

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type panicHandler struct {
	*SlogResponseHandler
	next http.Handler
}

func NewPanicHandler(next http.Handler, withOptions ...Option) (http.Handler, error) {
	o := &options{}
	err := WithOptions(append(withOptions,
		func(o *options) (err error) {
			if next == nil {
				return errors.New("cannot use a <nil> http.Handler")
			}
			if o.logger == nil {
				o.logger = slog.Default()
			}
			if o.successLevel == nil {
				if err = WithSuccessLevel(slog.LevelDebug)(o); err != nil {
					return err
				}
			}
			if o.errorLevel == nil {
				if err = WithErrorLevel(slog.LevelError * 2)(o); err != nil {
					return err
				}
			}
			if o.encoder == nil {
				if err = WithDefaultErrorTemplate()(o); err != nil {
					return err
				}
			}
			return nil
		},
	)...)(o)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize an structured logging panic handler: %w", err)
	}

	return &panicHandler{
		SlogResponseHandler: &SlogResponseHandler{
			logger:       o.logger,
			successLevel: *o.successLevel,
			errorLevel:   *o.errorLevel,
			encoder:      o.encoder,
		},
		next: next,
	}, nil
}

func (p *panicHandler) ServeHTTP(
	w http.ResponseWriter, r *http.Request,
) {
	defer func() {
		if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
			p.SlogResponseHandler.HandleError(
				w, r, fmt.Errorf("panic: %+v", rvr))
		}
	}()

	p.next.ServeHTTP(w, r)
}
