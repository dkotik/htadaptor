package slogrh

import (
	_ "embed" // for default error template
	"errors"
	"fmt"
	"html/template"
	"log/slog"
)

//go:embed error.html
var DefaultErrorTemplate string

type options struct {
	logger       *slog.Logger
	successLevel *slog.Level
	errorLevel   *slog.Level
	encoder      ErrorEncoder
}

type Option func(*options) error

// WithOptions combines several [Option]s into one.
func WithOptions(withOptions ...Option) Option {
	return func(o *options) (err error) {
		for _, option := range withOptions {
			if option == nil {
				return errors.New("cannot use a <nil> option")
			}
			if err = option(o); err != nil {
				return err
			}
		}
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(o *options) error {
		if logger == nil {
			return errors.New("cannot use a <nil> structured logger")
		}
		if o.logger != nil {
			return errors.New("logger is already set")
		}
		o.logger = logger
		return nil
	}
}

func WithSuccessLevel(level slog.Level) Option {
	return func(o *options) error {
		if o.successLevel != nil {
			return errors.New("success log level is already set")
		}
		o.successLevel = &level
		return nil
	}
}

func WithDefaultSuccessLevelInfo() Option {
	return WithSuccessLevel(slog.LevelInfo)
}

func WithErrorLevel(level slog.Level) Option {
	return func(o *options) error {
		if o.errorLevel != nil {
			return errors.New("error log level is already set")
		}
		o.errorLevel = &level
		return nil
	}
}

func WithDefaultErrorLevelError() Option {
	return WithErrorLevel(slog.LevelError)
}

func WithErrorEncoder(e ErrorEncoder) Option {
	return func(o *options) error {
		if e == nil {
			return errors.New("cannot use a <nil> error encoder")
		}
		if o.encoder != nil {
			return errors.New("error encoder is already set")
		}
		o.encoder = e
		return nil
	}
}

func WithJSONEncoder() Option {
	return WithErrorEncoder(DefaultJSONErrorEncoder)
}

func WithErrorTemplate(t *template.Template) Option {
	return func(o *options) error {
		encoder, err := NewTemplateErrorEncoder(t)
		if err != nil {
			return err
		}
		return WithErrorEncoder(encoder)(o)
	}
}

func WithDefaultErrorTemplate() Option {
	return func(o *options) error {
		t, err := template.New("error").Parse(DefaultErrorTemplate)
		if err != nil {
			return fmt.Errorf("cannot compile default error template: %w", err)
		}
		return WithErrorTemplate(t)(o)
	}
}
