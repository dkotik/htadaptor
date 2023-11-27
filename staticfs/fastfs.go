package staticfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/slogrh"
)

type fastFileSystemFile struct {
	ContentType string
	Contents    []byte
}

func (f *fastFileSystemFile) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("content-type", f.ContentType)
	_, _ = io.Copy(w, bytes.NewReader(f.Contents))
}

// NewFastFileSystemFile serves a single file from memory without logging.
func NewFastFileSystemFile(contents []byte) http.Handler {
	return &fastFileSystemFile{
		ContentType: http.DetectContentType(contents),
		Contents:    contents,
	}
}

type fastFileSystem struct {
	index           map[string]*fastFileSystemFile
	responseHandler htadaptor.ResponseHandler
}

func (f *fastFileSystem) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) error {
	p := r.URL.Path
	file, ok := f.index[p]
	if !ok {
		return NewNotFoundError(p)
	}
	w.Header().Set("content-type", file.ContentType)
	if _, err := io.Copy(w, bytes.NewReader(file.Contents)); err != nil {
		return err
	}
	return f.responseHandler.HandleSuccess(w, r)
}

func (f *fastFileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f.ServeHyperText(w, r)
	if err != nil {
		f.responseHandler.HandleError(w, r, err)
	}
}

type fastFileSystemOptions struct {
	index                  map[string]*fastFileSystemFile
	responseHandler        htadaptor.ResponseHandler
	responseHandlerOptions []slogrh.Option
}

type FastFileSystemOption func(*fastFileSystemOptions) error

func NewFastFileSystem(withOptions ...FastFileSystemOption) (_ http.Handler, err error) {
	o := &fastFileSystemOptions{
		index: make(map[string]*fastFileSystemFile),
	}
	for _, option := range append(
		withOptions,
		func(o *fastFileSystemOptions) (err error) {
			if len(o.responseHandlerOptions) > 0 {
				if o.responseHandler != nil {
					return fmt.Errorf("option WithResponseHandler conflicts with %d response handler options; provide either a prepared response handler or options for preparing an slog one, but not both", len(o.responseHandlerOptions))
				}
				o.responseHandler, err = slogrh.New(o.responseHandlerOptions...)
				if err != nil {
					return err
				}
			}
			if o.responseHandler == nil {
				if err = WithDefaultFastFileSystemResponseHandler()(o); err != nil {
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
		index:           o.index,
		responseHandler: o.responseHandler,
	}, nil
}

func WithFastFileSystemFile(
	path string,
	contents []byte,
) FastFileSystemOption {
	return func(o *fastFileSystemOptions) error {
		if path == "" {
			return errors.New("cannot use an empty path")
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
			ContentType: http.DetectContentType(contents),
			Contents:    b,
		}
		return nil
	}
}

func WithFastFileSystemResponseHandler(h htadaptor.ResponseHandler) FastFileSystemOption {
	return func(o *fastFileSystemOptions) error {
		if h == nil {
			return errors.New("cannot use a <nil> response handler")
		}
		if o.responseHandler != nil {
			return errors.New("response handler is already set")
		}
		o.responseHandler = h
		return nil
	}
}

func WithDefaultFastFileSystemResponseHandler() FastFileSystemOption {
	return func(o *fastFileSystemOptions) error {
		handler, err := slogrh.New(
			slogrh.WithSuccessLevel(slog.LevelDebug),
		)
		if err != nil {
			return err
		}
		return WithFastFileSystemResponseHandler(handler)(o)
	}
}

// TODO: write out slogrh.Options for convenience.
