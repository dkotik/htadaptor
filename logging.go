package htadaptor

import (
	"log/slog"
	"net/http"
)

var voidLogger RequestLogger

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
	if voidLogger == nil {
		voidLogger = RequestLoggerFunc(func(r *http.Request, err error) {})
	}
	return voidLogger
}
