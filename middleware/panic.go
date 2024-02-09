package middleware

import (
	"log/slog"
	"net/http"
)

type panicHandler struct {
	next       http.Handler
	lastResort http.Handler
}

// NewPanic returns middleware that prevents service shutdown by
// recovering from panics. Then, it logs the recovery value and
// runs lastResort [http.Handler]. If lastResort is <nil>
// the service will panic again and shutdown.
func NewPanic(lastResort http.Handler) Middleware {
	return func(next http.Handler) http.Handler {
		return &panicHandler{
			next:       next,
			lastResort: lastResort,
		}
	}
}

func (p *panicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rvr := recover(); rvr != nil {
			slog.Default().ErrorContext(
				r.Context(),
				"request execution ended with panic",
				slog.Any("panic", rvr),
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				// TODO: add a stack trace
			)
			p.lastResort.ServeHTTP(w, r)
		}
	}()

	p.next.ServeHTTP(w, r)
}
