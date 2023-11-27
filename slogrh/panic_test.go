package slogrh

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPanicRecovery(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthcheck", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	ph, err := NewPanicHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	}))
	if err != nil {
		t.Fatal(err)
	}
	ph.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatal("status code was not set to 500")
	}

	// t.Fatal(rr.Body.String())
}
