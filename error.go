package htadaptor

import (
	"errors"
	"html/template"
	"net/http"
)

type ErrorHandler interface {
	HandleError(http.ResponseWriter, *http.Request, error) error
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error) error

func (e ErrorHandlerFunc) HandleError(w http.ResponseWriter, r *http.Request, err error) error {
	return e(w, r, err)
}

func NewErrorHandler(encoder Encoder) ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) error {
		w.WriteHeader(GetHyperTextStatusCode(err))
		return errors.Join(err, encoder.Encode(w, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}))
	}
}

func NewErrorHandlerFromTemplate(t *template.Template) ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) error {
		w.Header().Set("Content-Type", "text/html")
		code := GetHyperTextStatusCode(err)
		w.WriteHeader(code)
		return errors.Join(err, t.Execute(w, struct {
			Title   string
			Message string
		}{
			Title:   http.StatusText(code),
			Message: err.Error(),
		}))
	}
}

type Error interface {
	error
	HyperTextStatusCode() int
}

func GetHyperTextStatusCode(err error) int {
	var htError Error
	if errors.As(err, &htError) {
		return htError.HyperTextStatusCode()
	}
	return http.StatusInternalServerError
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

type DecodingError struct {
	error
}

func NewDecodingError(fromError error) *DecodingError {
	return &DecodingError{fromError}
}

func (e *DecodingError) Error() string {
	return "unable to decode request: " + e.error.Error()
}

func (e *DecodingError) Unwrap() error {
	return e.error
}

func (e *DecodingError) HyperTextStatusCode() int {
	return http.StatusUnprocessableEntity
}

type EncodingError struct {
	error
}

func NewEncodingError(fromError error) *EncodingError {
	return &EncodingError{fromError}
}

func (e *EncodingError) Error() string {
	return "unable to encode response: " + e.error.Error()
}

func (e *EncodingError) Unwrap() error {
	return e.error
}

func (e *EncodingError) HyperTextStatusCode() int {
	return http.StatusInternalServerError
}
