package tmodulator

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimingModulator(t *testing.T) {
	if testing.Short() {
		t.Skip("timing modulator is a longer test")
	}

	h := New(time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, "test request")
		if err != nil {
			t.Fatal(err)
		}
	}))

	from := time.Now()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if took := time.Now().Sub(from); took < time.Second/2 {
		t.Fatal("request executed faster than one second:", took)
	}
}
