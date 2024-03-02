package htadaptor

import (
	"io"
	"net/http"
	"reflect"
	"strings"
)

// MethodSwitch provides method selections for [NewMethodMux].
type MethodSwitch struct {
	Get    http.Handler
	Post   http.Handler
	Put    http.Handler
	Patch  http.Handler
	Delete http.Handler
	Head   http.Handler
}

func (ms *MethodSwitch) AllowedMethods() (methods []string) {
	if ms.Get != nil {
		methods = append(methods, http.MethodGet)
	}
	if ms.Post != nil {
		methods = append(methods, http.MethodPost)
	}
	if ms.Put != nil {
		methods = append(methods, http.MethodPut)
	}
	if ms.Patch != nil {
		methods = append(methods, http.MethodPatch)
	}
	if ms.Delete != nil {
		methods = append(methods, http.MethodDelete)
	}
	return
}

type methodMux struct {
	Get     http.Handler
	Post    http.Handler
	Put     http.Handler
	Patch   http.Handler
	Delete  http.Handler
	Head    http.Handler
	allowed string
}

// NewMethodMux returns a handler that is able to satisfy REST
// interface expectations. It does not modify response status codes,
// but they can be updated using [WithStatusCode] option for
// individual handlers.
func NewMethodMux(ms *MethodSwitch) http.Handler {
	if ms == nil {
		return &getPostMux{}
	}
	allowed := ms.AllowedMethods()
	if len(allowed) == 2 && reflect.DeepEqual(allowed, []string{"GET", "POST"}) {
		return &getPostMux{
			Get:  ms.Get,
			Post: ms.Post,
		}
	}
	if ms.Head == nil {
		ms.Head = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	}
	return &methodMux{
		Get:     ms.Get,
		Post:    ms.Post,
		Put:     ms.Put,
		Patch:   ms.Patch,
		Delete:  ms.Delete,
		Head:    ms.Head,
		allowed: strings.Join(append(allowed, http.MethodHead), ", "),
	}
}

func (m *methodMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method { // http.Request ALWAYS has a method
	case http.MethodGet:
		if m.Get != nil {
			m.Get.ServeHTTP(w, r)
			return
		}
	case http.MethodPost:
		if m.Post != nil {
			m.Post.ServeHTTP(w, r)
			return
		}
	case http.MethodPut:
		if m.Put != nil {
			m.Put.ServeHTTP(w, r)
			return
		}
	case http.MethodPatch:
		if m.Patch != nil {
			m.Patch.ServeHTTP(w, r)
			return
		}
	case http.MethodDelete:
		if m.Delete != nil {
			m.Delete.ServeHTTP(w, r)
			return
		}
	case http.MethodOptions:
		w.Header().Set("Allow", m.allowed)
		return
	case http.MethodHead:
		m.Head.ServeHTTP(w, r)
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = io.WriteString(w, http.StatusText(http.StatusMethodNotAllowed))
}

type getPostMux struct {
	Get  http.Handler
	Post http.Handler
}

func (m *getPostMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method { // http.Request ALWAYS has a method
	case http.MethodGet:
		m.Get.ServeHTTP(w, r)
		return
	case http.MethodPost:
		m.Post.ServeHTTP(w, r)
		return
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, HEAD")
		return
	case http.MethodHead:
		return // no operation
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = io.WriteString(w, http.StatusText(http.StatusMethodNotAllowed))
}
