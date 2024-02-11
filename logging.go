package htadaptor

import (
	"log/slog"
	"net/http"
)

type Logger interface {
	LogRequest(*http.Request, error)
}

type LoggerFunc func(*http.Request, error)

func (f LoggerFunc) LogRequest(r *http.Request, err error) {
	f(r, err)
}

type SlogLogger struct {
	Logger  *slog.Logger
	Success slog.Level
	Error   slog.Level
}

func (s *SlogLogger) LogRequest(r *http.Request, err error) {
	if err == nil {
		s.Logger.Log(
			r.Context(),
			s.Success,
			"HTTP request served",
			slog.String("client_address", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.URL.String()),
		)
		return
	}
	s.Logger.Log(
		r.Context(),
		s.Error,
		"HTTP request failed",
		slog.Any("error", err),
		slog.String("client_address", r.RemoteAddr),
		slog.String("method", r.Method),
		slog.String("host", r.Host),
		slog.String("path", r.URL.String()),
	)
}
