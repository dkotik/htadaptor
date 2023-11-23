package htadaptor

import (
	"log/slog"
	"net/http"
)

type RequestLogger interface {
	LogRequest(*http.Request, error)
}

type RequestLoggerFunc func(*http.Request, error)

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
			"completed HTTP request via void adaptor",
			slog.String("client_address", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.String()),
		)
	} else {
		l.ErrorContext(
			r.Context(),
			err.Error(),
			slog.String("client_address", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.String()),
		)
	}
}

func NewRequestLogger(logger *slog.Logger, successLevel slog.Leveler) RequestLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &slogLogger{
		Logger:       logger,
		successLevel: successLevel.Level(),
	}
}

func NewVoidLogger() RequestLogger {
	return RequestLoggerFunc(func(r *http.Request, err error) {})
}
