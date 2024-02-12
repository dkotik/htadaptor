package staticfs

import (
	"log/slog"
	"net/http"

	"github.com/dkotik/htadaptor"
)

func (fs *FS) ServeHyperText(
	w http.ResponseWriter,
	r *http.Request,
) (err error) {
	real, ok := fs.index[r.URL.Path]
	if !ok {
		return htadaptor.NewNotFoundError(r.URL.Path)
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
