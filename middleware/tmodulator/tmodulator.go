/*
Package tmodulator delays HTTP responses to protects wrapped endpoints
from timing attacks.
*/
package tmodulator

import (
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/dkotik/htadaptor"
)

type timingModulator struct {
	next  http.Handler
	delay func(time.Time) time.Duration
}

func (t *timingModulator) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	from := time.Now()
	rec := httptest.NewRecorder()
	t.next.ServeHTTP(rec, r)
	time.Sleep(t.delay(from))

	// replay the recorded response
	result := rec.Result()
	for k, v := range result.Header {
		w.Header().Set(k, strings.Join(v, ","))
	}
	w.WriteHeader(result.StatusCode)
	_, _ = io.Copy(w, result.Body)
}

func New(normalizedTiming time.Duration) htadaptor.Middleware {
	if normalizedTiming < time.Millisecond {
		panic("cannot modulate time for less than a millisecond")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(next http.Handler) http.Handler {
		return &timingModulator{
			next: next,
			delay: func(from time.Time) time.Duration {
				jitter := 1 + r.NormFloat64()*0.5 // range from 0.5 to 1.5
				delay := time.Duration(float64(normalizedTiming) * jitter)
				// fmt.Println(delay)
				// panic(delay)
				return time.Now().Add(delay).Sub(from)
			},
		}
	}
}
