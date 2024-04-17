package htadaptor

import (
	"fmt"
	"io"
	"log/slog"
	"sync"

	_ "embed" // for error template

	"errors"
	"html/template"
	"net/http"
)

// Error extends the [error] interface to include HTTP status code.
type Error interface {
	error
	HyperTextStatusCode() int
}

var (
	//go:embed error.html
	defaultErrorTemplateSource []byte
	defaultErrorTemplateSetup  sync.Once
	defaultErrorTemplate       *template.Template
)

func DefaultErrorTemplate() *template.Template {
	defaultErrorTemplateSetup.Do(func() {
		t, err := template.New("error").Parse(string(defaultErrorTemplateSource))
		if err != nil {
			panic(fmt.Errorf("could not pase default template: %w", err))
		}
		defaultErrorTemplate = t
	})
	return defaultErrorTemplate
}

var _ slog.LogValuer = (*NotFoundError)(nil)

type ErrorHandler interface {
	HandleError(http.ResponseWriter, *http.Request, error) error
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error) error

func (e ErrorHandlerFunc) HandleError(w http.ResponseWriter, r *http.Request, err error) error {
	return e(w, r, err)
}

func NewErrorHandler(encoder Encoder) ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) error {
		return errors.Join(err, encoder.Encode(w, r, GetHyperTextStatusCode(err), struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}))
	}
}

type ErrorMessage struct {
	StatusCode int
	Title      string
	Message    string
}

func (e *ErrorMessage) Render(w io.Writer) error {
	return DefaultErrorTemplate().Execute(w, e)
}

func NewErrorHandlerFromTemplate(t *template.Template) ErrorHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) error {
		w.Header().Set("content-type", "text/html")
		code := GetHyperTextStatusCode(err)
		w.WriteHeader(code)
		return errors.Join(err, t.Execute(w, &ErrorMessage{
			StatusCode: code,
			Title:      http.StatusText(code),
			Message:    err.Error(),
		}))
	}
}

func GetHyperTextStatusCode(err error) int {
	var htError Error
	if errors.As(err, &htError) {
		return htError.HyperTextStatusCode()
	}
	return http.StatusInternalServerError
}

type InvalidRequestError struct {
	error
}

type DecodingError struct {
	error
}

func NewDecodingError(fromError error) Error {
	underlying, ok := fromError.(Error)
	if ok {
		// the underlying error has a more precise HTTP
		// status code than http.StatusUnprocessableEntity
		// which will be assigned by [DecodingError]
		return underlying
	}
	return &DecodingError{fromError}
}

func (e *DecodingError) Error() string {
	return e.error.Error()
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

func NewEncodingError(fromError error) Error {
	underlying, ok := fromError.(Error)
	if ok {
		return underlying
	}
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

type NotFoundError struct {
	path string
}

func NewNotFoundError(p string) *NotFoundError {
	return &NotFoundError{path: p}
}

func (e *NotFoundError) Error() string {
	return "Not Found"
}

func (e *NotFoundError) HyperTextStatusCode() int {
	return http.StatusNotFound
}

func (e *NotFoundError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("path", e.path),
	)
}
