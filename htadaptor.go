/*
Package htadaptor provides generic domain logic adaptors for HTTP handlers. Adaptors come in three flavors:

 1. UnaryFunc: func(context, inputStruct) (outputStruct, error)
 2. NullaryFunc: func(context) (outputStruct, error)
 3. VoidFunc: func(context, inputStruct) error

Validation errors are decorated with the correct [http.StatusUnprocessableEntity] status code.
*/
package htadaptor

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"reflect"

	"github.com/dkotik/htadaptor/reflectd"
)

// Validatable constrains a domain request. Validation errors are
// wrapped as [InvalidRequestError] by the adapter. [context.Context]
// is essential for passing locale information that can be
// retrieved using [LanguageFromContext] inside the validation
// method and other similar uses.
//
// Validatable is a generic interface that requires that the underlying
// type T is a pointer type. This is enforced by the [*T] constraint and
// is critical for type reflection to work correctly on struct fields.
type Validatable[T any] interface {
	*T
	Validate(context.Context) error
}

type Decoder interface {
	Decode(any, *http.Request) error
}

type DecoderFunc func(any, *http.Request) error

func (f DecoderFunc) Decode(v any, r *http.Request) error {
	return f(v, r)
}

type Encoder interface {
	Encode(http.ResponseWriter, *http.Request, int, any) error
}

type EncoderFunc func(http.ResponseWriter, *http.Request, int, any) error

func (f EncoderFunc) Encode(w http.ResponseWriter, r *http.Request, code int, v any) error {
	return f(w, r, code, v)
}

var JSONEncoder = EncoderFunc(
	func(w http.ResponseWriter, r *http.Request, code int, v any) error {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(code)
		return json.NewEncoder(w).Encode(v)
	},
)

type templateEncoder struct {
	*template.Template
}

func NewTemplateEncoder(t *template.Template) Encoder {
	return &templateEncoder{t}
}

func (e *templateEncoder) Encode(w http.ResponseWriter, r *http.Request, code int, v any) error {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	return e.Template.Execute(w, v)
}

// type TemplateMatcher interface {
// 	MatchTemplate(any) *template.Template
// }

// type templateMatchingEncoder struct {
// 	TemplateMatcher
// }

// func NewTemplateMatchingEncoder(m TemplateMatcher) Encoder {
// 	return &templateMatchingEncoder{m}
// }

// func (e *templateMatchingEncoder) Encode(w http.ResponseWriter, r *http.Request, code int, v any) error {
// 	t := e.MatchTemplate(v)
// 	if t == nil {
// 		return fmt.Errorf("no template matched for %v (%T)", v, v)
// 	}
// 	w.Header().Set("content-type", "text/html; charset=utf-8")
// 	w.WriteHeader(code)
// 	return t.Execute(w, v)
// }

// Must panics if an [http.Handler] was created with an error.
func Must(h http.Handler, err error) http.Handler {
	if err != nil {
		panic(err)
	}
	return h
}

// Middleware modifies an [http.Handler].
type Middleware func(http.Handler) http.Handler

// Apply wraps an [http.Handler] into [Middleware] in reverse order.
func ApplyMiddleware(h http.Handler, mws ...Middleware) http.Handler {
	if h == nil {
		panic("cannot use <nil> handler")
	}
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

type Adaptor struct {
	options []Option
}

func (a Adaptor) initialize(withOptions []Option) (o *options, err error) {
	o = &options{
		StatusCode: http.StatusOK,
	}
	if err = WithOptions(a.options...)(o); err != nil {
		return o, fmt.Errorf("invalid core option: %w", err)
	}
	err = WithOptions(append(
		withOptions,
		func(o *options) (err error) {
			if o.Encoder == nil {
				if err = WithDefaultEncoder()(o); err != nil {
					return err
				}
			}
			if o.Decoder == nil {
				if len(o.DecoderOptions) == 0 {
					if err = WithDefaultDecoder()(o); err != nil {
						return err
					}
				} else {
					o.Decoder, err = reflectd.NewDecoder(o.DecoderOptions...)
					if err != nil {
						return err
					}
				}
			}
			if o.ErrorHandler == nil {
				if err = WithDefaultErrorHandler()(o); err != nil {
					return err
				}
			}
			return nil
		},
	)...)(o)
	return o, err
}

// New initializes a generic hyper text [Adaptor] with the given options
// that are used as the default options for the adapted domain calls.
func New(withOptions ...Option) Adaptor {
	// return Adaptor{
	// 	options: append([]Option{
	// 		unpanic.New(),
	// 	}, withOptions...),
	// }
	return Adaptor{
		options: withOptions,
	}
}

func ensureValuePointer(v any) any {
	rv := reflect.ValueOf(v)

	// If it is already a pointer, return it as-is
	if rv.Kind() == reflect.Ptr {
		return v
	}

	// If it's not a pointer, create a new pointer and populate it
	// ptr := reflect.New(rv.Type())
	// ptr.Elem().Set(rv)
	// return ptr.Interface()
	return reflect.PointerTo(rv.Type())
}

func isPointer(v any) bool {
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr
}
