/*
Package htadaptor provides generic domain logic adaptors for HTTP handlers. Adaptors come in three flavors:

1. UnaryFunc: func(context, inputStruct) (outputStruct, error)
2. NullaryFunc: func(context) (outputStruct, error)
3. VoidFunc: func(context, inputStruct) error

Each input requires implementation of [Validatable] for safety. Validation errors are decorated with the correct [http.StatusUnprocessableEntity] status code.
*/
package htadaptor

import (
	"net/http"
)

// func New(withOptions ...Option) (func(any) (http.Handler, error), error) {
// 	o := &options{}
// 	err := WithOptions(append(
// 		withOptions,
// 		func(o *options) (err error) {
// 			defer func() {
// 				if err != nil {
// 					err = fmt.Errorf("unable to apply default option: %w", err)
// 				}
// 			}()
//
// 			if o.Encoder == nil {
// 				if err = WithDefaultEncoder()(o); err != nil {
// 					return err
// 				}
// 			}
// 			if o.Decoder == nil {
// 				if err = WithDefaultDecoder()(o); err != nil {
// 					return err
// 				}
// 			}
// 			if o.ErrorHandler == nil {
// 				if err = WithDefaultErrorHandler()(o); err != nil {
// 					return err
// 				}
// 			}
// 			if o.Logger == nil {
// 				if err = WithDefaultLogger()(o); err != nil {
// 					return err
// 				}
// 			}
// 			return nil
// 		},
// 	)...)(o)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to initialize a generic adaptor: %w", err)
// 	}
//
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

type Error interface {
	error
	HyperTextStatusCode() int
}

type ErrorHandler interface {
	HandleError(http.ResponseWriter, *http.Request, error)
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

func (f ErrorHandlerFunc) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}

type Encoder interface {
	Encode(http.ResponseWriter, any) error
}

type EncoderFunc func(http.ResponseWriter, any) error

func (f EncoderFunc) Encode(w http.ResponseWriter, v any) error {
	return f(w, v)
}

type Decoder interface {
	Decode(any, *http.Request) error
}

type StringExtractor func(*http.Request) (string, error)

func Must(h http.Handler, err error) http.Handler {
	if err != nil {
		panic(err)
	}
	return h
}
