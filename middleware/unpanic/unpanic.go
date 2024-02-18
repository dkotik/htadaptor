/*
Package unpanic provides [htadaptor.Middleware] that gracefully recovers
from request panics.
*/
package unpanic

import (
	"log/slog"
	"net/http"

	"github.com/dkotik/htadaptor"
)

type panicHandler struct {
	next       http.Handler
	lastResort http.Handler
}

// New returns middleware that prevents service shutdown by
// recovering from panics. Then, it logs the recovery value and
// runs lastResort [http.Handler]. If lastResort is <nil>
// the service will panic again and shutdown.
func New(lastResort http.Handler) htadaptor.Middleware {
	if lastResort == nil {
		panic("cannot use a <nil> last resort handler")
	}

	return func(next http.Handler) http.Handler {
		if next == nil {
			panic("cannot use a <nil> next handler")
		}

		return &panicHandler{
			next:       next,
			lastResort: lastResort,
		}
	}
}

// ServeHTTP satisfies [http.Handler] interface.
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
