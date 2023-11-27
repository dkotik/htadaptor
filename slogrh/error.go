package slogrh

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
)

var (
	DefaultJSONErrorEncoder = ErrorEncoderFunc(func(w http.ResponseWriter, err error) error {
		message := NewErrorMessage(err)
		w.WriteHeader(message.StatusCode)
		return json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: message.Message,
		})
	})
)

type Error interface {
	error
	HyperTextStatusCode() int
}

type ErrorMessage struct {
	Message    string
	StatusCode int
}

func NewErrorMessage(err error) ErrorMessage {
	// if err == nil {
	// 	return ErrorMessage{
	// 		Message:    "<nil> error",
	// 		StatusCode: http.StatusInternalServerError,
	// 	}
	// }
	var htError Error
	if errors.As(err, &htError) {
		return ErrorMessage{
			Message:    err.Error(),
			StatusCode: htError.HyperTextStatusCode(),
		}
	}
	return ErrorMessage{
		Message:    err.Error(),
		StatusCode: http.StatusInternalServerError,
	}
}

type ErrorEncoder interface {
	EncodeError(http.ResponseWriter, error) error
}

type ErrorEncoderFunc func(http.ResponseWriter, error) error

func (f ErrorEncoderFunc) EncodeError(w http.ResponseWriter, err error) error {
	return f(w, err)
}

func NewTemplateErrorEncoder(t *template.Template) (ErrorEncoder, error) {
	if t == nil {
		return nil, errors.New("cannot use a <nil> error template")
	}
	return ErrorEncoderFunc(func(w http.ResponseWriter, err error) error {
		message := NewErrorMessage(err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(message.StatusCode)
		return t.Execute(w, message)
	}), nil
}
