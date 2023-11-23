package htadaptor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type fastFileSystemFile struct {
	ContentType string
	Contents    []byte
}

type fastFileSystem struct {
	index        map[string]*fastFileSystemFile
	errorHandler ErrorHandler
	logger       RequestLogger
}

func (f *fastFileSystem) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	p := r.URL.Path
	file, ok := f.index[p]
	if !ok {
		return &NotFoundError{path: p}
	}
	w.Header().Set("content-type", file.ContentType)
	_, err := io.Copy(w, bytes.NewReader(file.Contents))
	return err
}

func (f *fastFileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f.ServeHyperText(w, r)
	f.logger.LogRequest(r, err)
	if err != nil {
		f.errorHandler.HandleError(w, r, err)
	}
}

type fastFileSystemOptions struct {
	index        map[string]*fastFileSystemFile
	errorHandler ErrorHandler
	logger       RequestLogger
}

type FastFileSystemOption func(*fastFileSystemOptions) error

func NewFastFileSystem(withOptions ...FastFileSystemOption) (_ http.Handler, err error) {
	o := &fastFileSystemOptions{
		index: make(map[string]*fastFileSystemFile),
	}
	for _, option := range append(
		withOptions,
		func(o *fastFileSystemOptions) (err error) {
			if o.errorHandler == nil {
				if err = WithDefaultFastFileSystemErrorHandler()(o); err != nil {
					return err
				}
			}
			if o.logger == nil {
				if err = WithDefaultFastFileSystemLogger()(o); err != nil {
					return err
				}
			}
			return nil
		},
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("unable to initialize fast file system adaptor: %w", err)
		}
	}

	return &fastFileSystem{
		index:        o.index,
		errorHandler: o.errorHandler,
		logger:       o.logger,
	}, nil
}

func WithFastFileSystemFile(
	path string,
	contentType string,
	contents []byte,
) FastFileSystemOption {
	return func(o *fastFileSystemOptions) error {
		if path == "" {
			return errors.New("cannot use an empty path")
		}
		if contentType == "" {
			return fmt.Errorf("cannot use an empty content type: %s", path)
		}
		if len(contents) < 1 {
			return fmt.Errorf("cannot use empty file contents: %s", path)
		}
		if _, ok := o.index[path]; ok {
			return fmt.Errorf("file path is already set: %s", path)
		}
		b := make([]byte, len(contents))
		copy(b, contents)
		o.index[path] = &fastFileSystemFile{
			ContentType: contentType,
			Contents:    b,
		}
		return nil
	}
}

func WithFastFileSystemErrorHandler(e ErrorHandler) FastFileSystemOption {
	return func(o *fastFileSystemOptions) error {
		if e == nil {
			return errors.New("cannot use a <nil> error handler")
		}
		if o.errorHandler != nil {
			return errors.New("error handler is already set")
		}
		o.errorHandler = e
		return nil
	}
}

func WithFastFileSystemErrorHandlerFunc(f func(http.ResponseWriter, *http.Request, error)) FastFileSystemOption {
	return WithFastFileSystemErrorHandler(
		ErrorHandlerFunc(f),
	)
}

func WithDefaultFastFileSystemErrorHandler() FastFileSystemOption {
	return WithFastFileSystemErrorHandlerFunc(defaultErrorHandler)
}

func WithFastFileSystemLogger(l RequestLogger) FastFileSystemOption {
	return func(o *fastFileSystemOptions) error {
		if l == nil {
			return errors.New("cannot use a <nil> request logger")
		}
		if o.logger != nil {
			return errors.New("request logger is already set")
		}
		o.logger = l
		return nil
	}
}

func WithFastFileSystemSlogLogger(
	l *slog.Logger,
	successLevel slog.Leveler,
) FastFileSystemOption {
	return func(o *fastFileSystemOptions) error {
		if l == nil {
			return errors.New("cannot use a <nil> structured logger")
		}
		return WithFastFileSystemLogger(NewRequestLogger(l, successLevel))(o)
	}
}

func WithDefaultFastFileSystemLogger() FastFileSystemOption {
	return WithFastFileSystemLogger(NewVoidLogger())
}
