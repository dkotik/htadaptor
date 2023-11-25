package staticfs

import (
	"net/http"

	"log/slog"
)

var _ slog.LogValuer = (*NotFoundError)(nil)

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

func (fs *FS) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	real, ok := fs.index[r.URL.Path]
	if !ok {
		return NewNotFoundError(r.URL.Path)
	}
	r.URL.Path = real // TODO: not kosher.
	// r.URL.Path = "main.go"
	fs.source.ServeHTTP(w, r)
	return nil
}

func (fs *FS) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := fs.ServeHyperText(w, r)
	if err == nil {
		return
	}
	// var httpError Error
	// if errors.As(err, &httpError) {
	// 	msg := err.Error()
	// 	http.Error(w, msg, httpError.HyperTextStatusCode())
	// 	slog.Log(
	// 		r.Context(),
	// 		slog.LevelWarn,
	// 		msg,
	// 		slog.Any("error", err),
	// 	)
	// 	return
	// }
	http.Error(w, err.Error(), http.StatusInternalServerError)
	slog.Log(
		r.Context(),
		slog.LevelError,
		err.Error(),
		slog.String("path", r.URL.Path),
	)
}
