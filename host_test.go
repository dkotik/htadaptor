package htadaptor

import (
	"io"
	"net/http"
	"testing"
)

func TestHostMuxCreation(t *testing.T) {
	testHostHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello world!")
	})

	mux := NewHostMux(map[string]http.Handler{
		"one":   testHostHandler,
		"two":   testHostHandler,
		"three": testHostHandler,
	})
	_, ok := mux.(*listHostMux)
	if !ok {
		t.Fatal("created mux is not a listHostMux")
	}

	mux = NewHostMux(map[string]http.Handler{
		"one":   testHostHandler,
		"two":   testHostHandler,
		"three": testHostHandler,
		"4":     testHostHandler,
		"5":     testHostHandler,
		"6":     testHostHandler,
		"7":     testHostHandler,
		"8":     testHostHandler,
		"9":     testHostHandler,
	})

	_, ok = mux.(mapHostMux)
	if !ok {
		t.Fatal("created mux is not a mapHostMux")
	}
}
