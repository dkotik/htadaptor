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

	hosts := []HostMuxAssociation{
		{Name: "one", Handler: testHostHandler},
		{Name: "two", Handler: testHostHandler},
		{Name: "three", Handler: testHostHandler},
	}

	mux, err := NewHostMux(hosts...)
	_, ok := mux.(listHostMux)
	if !ok {
		t.Fatal("created mux is not a listHostMux")
	}
	if err != nil {
		t.Fatal(err)
	}

	hosts = append(hosts, []HostMuxAssociation{
		{Name: "1", Handler: testHostHandler},
		{Name: "2", Handler: testHostHandler},
		{Name: "3", Handler: testHostHandler},
		{Name: "4", Handler: testHostHandler},
		{Name: "5", Handler: testHostHandler},
		{Name: "6", Handler: testHostHandler},
		{Name: "7", Handler: testHostHandler},
	}...)

	mux, err = NewHostMux(hosts...)
	if err != nil {
		t.Fatal(err)
	}

	_, ok = mux.(mapHostMux)
	if !ok {
		t.Fatal("created mux is not a mapHostMux")
	}
}
