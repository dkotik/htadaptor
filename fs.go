package htadaptor

import (
	"log/slog"
	"net/http"
)

var _ slog.LogValuer = (*NotFoundError)(nil)

type NotFoundError struct {
	path string
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
