package staticfs

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
)

type FS struct {
	index  map[string]string
	source http.Handler
}

func New(withOptions ...Option) (_ *FS, err error) {
	o := &options{Index: make(map[string]string)}
	for _, option := range append(
		withOptions,
		func(o *options) error { // populate index
			if o.FileSystem == nil {
				return errors.New("file system is required")
			}
			if len(o.Translators) == 0 {
				if err := WithPathTranslators(
					func(real string) (external string, accept bool, err error) {
						return "/" + real, true, nil
					},
				)(o); err != nil {
					return err
				}
			}

			if err = fs.WalkDir(o.FileSystem, ".",
				func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil // skip directories
					}

					external := path
					accept := false
					for _, translator := range o.Translators {
						external, accept, err = translator(external)
						if err != nil {
							return err
						}
						if !accept {
							return nil // skip, choice of the translator
						}
						if err = WithPath(external, path)(o); err != nil {
							return err
						}
					}
					return nil
				},
			); err != nil {
				return fmt.Errorf("cannot index files from the file system: %w", err)
			}
			return nil
		},
	) {
		if err = option(o); err != nil {
			return nil, fmt.Errorf("cannot create static file system: %w", err)
		}
	}

	return &FS{
		index:  o.Index,
		source: http.FileServer(http.FS(o.FileSystem)),
	}, nil
}

func (fs *FS) String() string {
	return fmt.Sprintf("%+v", fs.index)
}
