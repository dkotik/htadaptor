package staticfs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type PathTranslator func(real string) (external string, accept bool, err error)

type options struct {
	Index       map[string]string
	Translators []PathTranslator
	FileSystem  fs.FS
}

type Option func(*options) error

func WithPath(external, real string) Option {
	return func(o *options) error {
		current, ok := o.Index[external]
		if ok {
			return fmt.Errorf("request path %q already points to %q", external, current)
		}
		o.Index[external] = real
		return nil
	}
}

func WithPathTranslators(ts ...PathTranslator) Option {
	return func(o *options) error {
		if len(ts) == 0 {
			return errors.New("empty path translator list")
		}
		for i, translator := range ts {
			if translator == nil {
				return fmt.Errorf("path translator list entry #%d is <nil>", i)
			}
		}
		o.Translators = append(o.Translators, ts...)
		return nil
	}
}

func WithFileSystem(fs fs.FS) Option {
	return func(o *options) (err error) {
		if fs == nil {
			return errors.New("unable to use a <nil> file system")
		}
		if o.FileSystem != nil {
			return errors.New("file system is already set")
		}
		o.FileSystem = fs
		return nil
	}
}

func WithDirectory(p string) Option {
	return WithFileSystem(os.DirFS(p))
}
