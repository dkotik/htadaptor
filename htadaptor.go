/*
Package htadaptor provides generic domain logic adaptors for HTTP handlers. Adaptors come in three flavors:

 1. UnaryFunc: func(context, inputStruct) (outputStruct, error)
 2. NullaryFunc: func(context) (outputStruct, error)
 3. VoidFunc: func(context, inputStruct) error

Each input requires implementation of [Validatable] for safety. Validation errors are decorated with the correct [http.StatusUnprocessableEntity] status code.
*/
package htadaptor

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
)

// func New(withOptions ...Option) (func(any) (http.Handler, error), error) {
// 	return func(domainCall any) (http.Handler, error) {
// 		funcType, err := Detect(domainCall)
// 		if err != nil {
// 			return nil, err
// 		}
// 		switch funcType {
// 		case FuncTypeUnary:
// 			return &UnaryFuncAdaptor{
// 				domainCall:   domainCall,
// 				decoder:      o.Decoder,
// 				encoder:      o.Encoder,
// 				errorHandler: o.ErrorHandler,
// 				logger:       o.Logger,
// 			}, nil
// 		default:
// 			return nil, fmt.Errorf("unknown domain function type: %d", funcType)
// 		}
// 	}, nil
// }

// Validatable constrains a domain request. Validation errors are
// wrapped as [InvalidRequestError] by the adapter. [context.Context]
// is essential for passing locale information that can be
// retrieved using [LanguageFromContext] inside the validation
// method and other similar uses.
type Validatable[T any] interface {
	*T
	Validate(context.Context) error
}

type Decoder interface {
	Decode(any, *http.Request) error
}

type DecoderFunc func(any, *http.Request) error

func (f DecoderFunc) Decoder(v any, r *http.Request) error {
	return f(v, r)
}

type Encoder interface {
	ContentType() string
	Encode(http.ResponseWriter, *http.Request, any) error
}

func setEncoderContentType(w http.ResponseWriter, e Encoder) {
	w.Header().Set("content-type", e.ContentType())
}

type JSONEncoder struct{}

func (e *JSONEncoder) ContentType() string {
	return "application/json"
}

func (e *JSONEncoder) Encode(w http.ResponseWriter, r *http.Request, v any) error {
	return json.NewEncoder(w).Encode(v)
}

type templateEncoder struct {
	*template.Template
}

func (e *templateEncoder) ContentType() string {
	return "text/html"
}

func (e *templateEncoder) Encode(w http.ResponseWriter, r *http.Request, v any) error {
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	}
	return e.Template.Execute(w, v)
}

func NewTemplateEncoder(t *template.Template) Encoder {
	return &templateEncoder{t}
}

// Must panics if an [http.Handler] was created with an error.
func Must(h http.Handler, err error) http.Handler {
	if err != nil {
		panic(err)
	}
	return h
}

type statusCodeEncoder struct {
	Encoder
	statusCode int
}

// NewStatusCodeEncoder returns an encoder that writes header
// first with a given status code before calling the next
// encoder. This replaces the default [http.StatusOK]
// with a different status code like [http.StatusCreated].
func NewStatusCodeEncoder(
	next Encoder,
	statusCode int,
) Encoder {
	return &statusCodeEncoder{
		Encoder:    next,
		statusCode: statusCode,
	}
}

func (e *statusCodeEncoder) Encode(
	w http.ResponseWriter,
	r *http.Request,
	v any,
) error {
	w.WriteHeader(e.statusCode)
	return e.Encoder.Encode(w, r, v)
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
