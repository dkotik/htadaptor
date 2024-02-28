/*
Package ctxlogger provides an [htadaptor.Middleware] that injects
a given logger into context so that it can be recovered
later in the stack for use.
*/
package ctxlogger

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dkotik/htadaptor"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

func FromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(contextKey).(*slog.Logger)
	if ok {
		return logger
	}
	return slog.Default()
}

func NewContext(parent context.Context, l *slog.Logger) context.Context {
	if l == nil {
		l = slog.Default()
	}
	return context.WithValue(parent, contextKey, l)
}

// New creates an [htadaptor.Middleware] that injects [slog.Logger]
// into request context so it can be recovered later using [FromContext].
func New(logger *slog.Logger) htadaptor.Middleware {
	return func(next http.Handler) http.Handler {
		if next == nil {
			panic("cannot use a <nil> next handler")
		}
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), logger)))
			},
		)
	}
}
