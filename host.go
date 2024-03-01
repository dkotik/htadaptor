package htadaptor

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"log/slog"
)

// HostMuxAssociation assigns a handler to a host for [NewHostMux]
// initialization.
//
// It is used instead of a `map[string]http.Handler` because
// the intended order of hosts is preserved in case
// the resulting handler uses a list implementation internally.
// Golang maps do not preserve the order of their keys or values.
type HostMuxAssociation struct {
	Name    string
	Handler http.Handler
}

// NewHostMux creates an [http.Handler] that multiplexes by
// [http.Request] host name.
func NewHostMux(hostHandlers ...HostMuxAssociation) (http.Handler, error) {
	if len(hostHandlers) == 0 {
		return nil, errors.New("cannot create host mux: no host associations provided")
	}

	handlers := make(mapHostMux)
	for _, association := range hostHandlers {
		if association.Name == "" {
			return nil, errors.New("cannot create host mux: cannot use an empty host name")
		}
		if association.Handler == nil {
			return nil, fmt.Errorf("cannot create host mux: host <%s> has a <nil> handler", association.Name)
		}
		if _, ok := handlers[association.Name]; ok {
			return nil, fmt.Errorf("cannot create host mux: host <%s> already has a handler", association.Name)
		}
		handlers[association.Name] = association.Handler
	}

	// mapHostMux will be faster than list at 8 entries
	if len(handlers) >= 8 {
		return handlers, nil
	}
	// preserves the original given order
	return listHostMux(hostHandlers), nil
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

type listHostMux []HostMuxAssociation

func (l listHostMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Hostname()
	for _, association := range l {
		if association.Name == name {
			association.Handler.ServeHTTP(w, r)
			return
		}
	}
	reportUnknownHost(w, r)
}
