package htadaptor

import (
	"log/slog"
	"net/http"
)

// voidLogger does not perform any logging.
var voidLogger RequestLogger

// RequestLogger creates records of each request and an error if one occured.
type RequestLogger interface {
	LogRequest(*http.Request, error)
}

// RequestLoggerFunc provides syntax sugar for satisfying [RequestLogger] interface with a function with the same signature.
type RequestLoggerFunc func(*http.Request, error)

// LogRequest redirects the interface responsility to its function's representation.
func (f RequestLoggerFunc) LogRequest(r *http.Request, err error) {
	f(r, err)
}

type slogLogger struct {
	*slog.Logger
	successLevel slog.Level
}

func (l *slogLogger) LogRequest(r *http.Request, err error) {
	if err == nil {
		l.Log(
			r.Context(),
			l.successLevel,
			"handled HTTP request",
			slog.String("client_address", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.String()),
		)
	} else {
		l.ErrorContext(
			r.Context(),
			"failed to handle HTTP request",
			slog.Any("error", err),
			slog.String("client_address", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.String()),
		)
	}
}

// NewRequestLogger records requests and errors into an [slog.Logger]. The success level specifies which [slog.Level] to use when no error occured.
func NewRequestLogger(logger *slog.Logger, successLevel slog.Leveler) RequestLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &slogLogger{
		Logger:       logger,
		successLevel: successLevel.Level(),
	}
}

// NewVoidLogger sets up a logger that performs no logging operations.
func NewVoidLogger() RequestLogger {
	if voidLogger == nil {
		voidLogger = RequestLoggerFunc(func(r *http.Request, err error) {})
	}
	return voidLogger
}
