package htadaptor

import (
	"fmt"
	"io"
	"net/http"

	"log/slog"
)

// NewHostMux creates an [http.Handler] that multiplexes by
// [http.Request] host name.
func NewHostMux(handlers map[string]http.Handler) http.Handler {
	// mapHostMux will be faster than list at 8 entries
	if len(handlers) >= 8 {
		for host, handler := range handlers {
			if handler == nil {
				panic(fmt.Errorf("HTTP handler for host %q is <nil>", host))
			}
		}
		return mapHostMux(handlers)
	}
	listHosts := make([]string, 0, len(handlers))
	listHandlers := make([]http.Handler, 0, len(handlers))
	for host, handler := range handlers {
		if handler == nil {
			panic(fmt.Errorf("HTTP handler for host %q is <nil>", host))
		}
		listHosts = append(listHosts, host)
		listHandlers = append(listHandlers, handler)
	}
	return &listHostMux{
		hosts:    listHosts,
		handlers: listHandlers,
	}
}

func reportUnknownHost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = io.WriteString(w, http.StatusText(http.StatusNotFound))
	slog.Default().WarnContext(
		r.Context(),
		"received unknown host request",
		slog.String("host", r.Host),
		// TODO: unwind the request fields using a helper method.
		// or should request be in context and injected using slog.Handler?
	)
}

type mapHostMux map[string]http.Handler

func (h mapHostMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Hostname()
	handler, ok := h[name]
	if !ok {
		reportUnknownHost(w, r)
		return
	}
	handler.ServeHTTP(w, r)
}

type listHostMux struct {
	hosts    []string
	handlers []http.Handler
}

func (l *listHostMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Hostname()
	for i, host := range l.hosts {
		if host == name {
			l.handlers[i].ServeHTTP(w, r)
			return
		}
	}
	reportUnknownHost(w, r)
}
