package htadaptor

import "net/http"

type Error interface {
	error
	HyperTextStatusCode() int
}

// Validatable constrains a domain request. Validation errors are wrapped as [InvalidRequestError] by the adapter.
type Validatable[T any] interface {
	*T
	Validate() error
}

type InvalidRequestError struct {
	error
}

func NewInvalidRequestError(fromError error) *InvalidRequestError {
	return &InvalidRequestError{fromError}
}

func (e *InvalidRequestError) Error() string {
	return "invalid request: " + e.error.Error()
}

func (e *InvalidRequestError) Unwrap() error {
	return e.error
}

func (e *InvalidRequestError) HyperTextStatusCode() int {
	return http.StatusUnprocessableEntity
}

type ErrorHandler interface {
	HandleError(http.ResponseWriter, *http.Request, error)
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

func (f ErrorHandlerFunc) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}

var DefaultErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	htError, ok := err.(Error)
	if ok {
		http.Error(w, err.Error(), htError.HyperTextStatusCode())
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}