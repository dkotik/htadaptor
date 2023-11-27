/*
Package slogrh handles domain logic reponses by logging them using [slog.Logger] and printing a simple error message.
*/
package slogrh

import (
	"fmt"
	"log/slog"
	"net/http"
)

// SlogResponseHandler logs successful and failed responses and displays the error using either JSON encoder or another encoder, which renders as "text/html" by default.
type SlogResponseHandler struct {
	logger       *slog.Logger
	successLevel slog.Level
	errorLevel   slog.Level
	encoder      ErrorEncoder
}

func New(withOptions ...Option) (*SlogResponseHandler, error) {
	o := &options{}
	err := WithOptions(append(withOptions,
		func(o *options) (err error) {
			if o.logger == nil {
				o.logger = slog.Default()
			}
			if o.successLevel == nil {
				if err = WithDefaultSuccessLevelInfo()(o); err != nil {
					return err
				}
			}
			if o.errorLevel == nil {
				if err = WithDefaultErrorLevelError()(o); err != nil {
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
		return nil, fmt.Errorf("failed to initialize an structured logging reponse handler: %w", err)
	}

	return &SlogResponseHandler{
		logger:       o.logger,
		successLevel: *o.successLevel,
		errorLevel:   *o.errorLevel,
		encoder:      o.encoder,
	}, nil
}

func (s *SlogResponseHandler) HandleSuccess(
	w http.ResponseWriter, r *http.Request,
) error {
	s.logger.Log(
		r.Context(),
		s.successLevel,
		"HTTP request served",
		slog.String("client_address", r.RemoteAddr),
		slog.String("method", r.Method),
		slog.String("host", r.Host),
		slog.String("path", r.URL.String()),
	)
	return nil
}

func (s *SlogResponseHandler) HandleError(
	w http.ResponseWriter, r *http.Request, err error,
) {
	s.logger.Log(
		r.Context(),
		s.errorLevel,
		"HTTP request failed",
		slog.Any("error", err),
		slog.String("client_address", r.RemoteAddr),
		slog.String("method", r.Method),
		slog.String("host", r.Host),
		slog.String("path", r.URL.String()),
	)
	if r.Header.Get("Content-Type") == "application/json" {
		DefaultJSONErrorEncoder.EncodeError(w, err)
	} else {
		s.encoder.EncodeError(w, err)
	}
}
