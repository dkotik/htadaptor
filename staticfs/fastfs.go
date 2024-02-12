package staticfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dkotik/htadaptor"
)

type FastFileSystemFile struct {
	contentType string
	contents    []byte
}

func (f *FastFileSystemFile) HandleError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) error {
	f.ServeHTTP(w, r)
	return nil
}

func (f *FastFileSystemFile) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("content-type", f.contentType)
	_, _ = io.Copy(w, bytes.NewReader(f.contents))
}

type fastFileSystem struct {
	Index       map[string]*FastFileSystemFile
	Fallthrough http.Handler
}

func (f *fastFileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	file, ok := f.Index[p]
	if !ok {
		f.Fallthrough.ServeHTTP(w, r)
		return
	}
	file.ServeHTTP(w, r)
}

// NewFastFileSystemFile serves a single file from memory without logging.
func NewFastFileSystemFile(contents []byte) *FastFileSystemFile {
	return &FastFileSystemFile{
		contentType: http.DetectContentType(contents),
		contents:    contents,
	}
}

type fastFileSystemOptions struct {
	Index       map[string]*FastFileSystemFile
	Fallthrough http.Handler
}

type FastFileSystemOption func(*fastFileSystemOptions) error

func NewFastFileSystem(withOptions ...FastFileSystemOption) (_ http.Handler, err error) {
	o := &fastFileSystemOptions{
		Index: make(map[string]*FastFileSystemFile),
	}
	for _, option := range append(
		withOptions,
		WithDefaultFastFileSystemFallthrough(),
		func(o *fastFileSystemOptions) (err error) {
			if len(o.Index) < 1 {
				return errors.New("provide at least one file")
			}
			return nil
		},
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("unable to initialize fast file system: %w", err)
		}
	}

	return &fastFileSystem{
		Index:       o.Index,
		Fallthrough: o.Fallthrough,
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
		if _, ok := o.Index[path]; ok {
			return fmt.Errorf("file path is already set: %s", path)
		}
		b := make([]byte, len(contents))
		copy(b, contents)
		o.Index[path] = &FastFileSystemFile{
			contentType: http.DetectContentType(contents),
			contents:    b,
		}
		return nil
	}
}

func WithFastFileSystemFallthrough(h http.Handler) FastFileSystemOption {
	return func(o *fastFileSystemOptions) error {
		if h == nil {
			return errors.New("cannot use a <nil> handler")
		}
		if o.Fallthrough != nil {
			return errors.New("fall through handler is already set")
		}
		o.Fallthrough = h
		return nil
	}
}

func WithDefaultFastFileSystemFallthrough() FastFileSystemOption {
	return func(o *fastFileSystemOptions) (err error) {
		if o.Fallthrough != nil {
			return nil
		}
		b := &bytes.Buffer{}
		if err = (&htadaptor.ErrorMessage{
			StatusCode: http.StatusNotFound,
			Title:      http.StatusText(http.StatusNotFound),
			Message:    "Requested file does not exist.",
		}).Render(b); err != nil {
			return fmt.Errorf("unable to render error message: %w", err)
		}
		return WithFastFileSystemFallthrough(
			&FastFileSystemFile{
				contentType: "text/html",
				contents:    b.Bytes(),
			},
		)(o)
	}
}
