package extract

import (
	"log/slog"
	"net/http"
)

// Error signals the internal failure of an adaptor operation.
type Error uint

const (
	ErrUnknown Error = iota
	ErrNoStringValue
	ErrUnsupportedMediaType
)

// Error satisfies [error] interface.
func (e Error) Error() string {
	switch e {
	case ErrNoStringValue:
		return "string value could not be recovered from request"
	case ErrUnsupportedMediaType:
		return http.StatusText(http.StatusUnsupportedMediaType)
	default:
		return "unknown extractor error"
	}
}

// HyperTextStatusCode satisfies [Error] interface.
func (e Error) HyperTextStatusCode() int {
	switch e {
	case ErrNoStringValue:
		return http.StatusUnprocessableEntity
	case ErrUnsupportedMediaType:
		return http.StatusUnsupportedMediaType
	default:
		return http.StatusInternalServerError
	}
}

// ReadLimitError indicates the failure to decode due to request intity
// containing more bytes than a handler is allowed to consume.
type ReadLimitError struct {
	limit int
}

func NewReadLimitError(limit int) *ReadLimitError {
	return &ReadLimitError{limit: limit}
}

func (e *ReadLimitError) Error() string {
	return http.StatusText(http.StatusRequestEntityTooLarge)
}

func (e *ReadLimitError) HyperTextStatusCode() int {
	return http.StatusRequestEntityTooLarge
}

func (e *ReadLimitError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("cause", "request body exceeded the adaptor read limit"),
		slog.Int("limit", e.limit),
	)
}

// // NoValueError indicates that none of the required values
// // from a list of possible value names were extracted.
// type NoValueError []string
//
// func (e NoValueError) Error() string {
// 	switch len(e) {
// 	case 0:
// 		return "request does not include required value"
// 	case 1:
// 		return fmt.Sprintf("request requires %q value", e[0])
// 	default:
// 		b := strings.Builder{}
// 		b.WriteString("request requires any of ")
// 		for _, name := range e {
// 			b.WriteRune('"')
// 			b.WriteString(name)
// 			b.WriteRune('"')
// 			b.WriteRune(',')
// 			b.WriteRune(' ')
// 		}
// 		return b.String()[:b.Len()-2] + " values"
// 	}
// }
