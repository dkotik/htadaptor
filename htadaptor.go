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

type Decoder interface {
	Decode(any, *http.Request) error
}

type DecoderFunc func(any, *http.Request) error

func (f DecoderFunc) Decoder(v any, r *http.Request) error {
	return f(v, r)
}

type Encoder interface {
	Encode(http.ResponseWriter, any) error
}

type EncoderFunc func(http.ResponseWriter, any) error

func (f EncoderFunc) Encode(w http.ResponseWriter, v any) error {
	return f(w, v)
}

type ResponseHandler interface {
	HandleSuccess(http.ResponseWriter, *http.Request) error
	HandleError(http.ResponseWriter, *http.Request, error)
}

/*
type silentResponseHandler struct{}

func (s *silentResponseHandler) HandleSuccess(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func (s *silentResponseHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
  // apply error encoder
}
*/

func Must(h http.Handler, err error) http.Handler {
	if err != nil {
		panic(err)
	}
	return h
}
