/*
Package session presents a lazy context that manages session state
with native key rotation support.
*/
package session

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Encoder interface {
	Encode(http.ResponseWriter, *http.Request, any) error
}

type Decoder interface {
	Decode(any, http.ResponseWriter, *http.Request) error
}

type Session interface {
	ID() string
	TraceID() string
	Get(string) any
	Set(string, any)
	Commit() error
	Rotate() error
}

func New(withOptions ...Option) (func(http.Handler) http.Handler, error) {
	options := &options{}
	var err error
	for _, option := range append(
		withOptions,
		WithDefaultName(),
		WithDefaultExpiry(),
		WithDefaultRotationContext(),
	) {
		if err = option(options); err != nil {
			return nil, fmt.Errorf("cannot create session middleware: %w", err)
		}
	}

	// TODO: instead of codec use a KV key repository.
	// TODO: add pooled decoder and encoder? no, the pool itself locks
	// a mutex!
	codec := newCodec()

	enc := &cookieEncoder{
		name:   options.Name,
		path:   "/",
		maxAge: int(options.Expiry.Seconds()),

		mu:      &sync.Mutex{},
		current: codec,
	}
	dec := &cookieDecoder{
		name: options.Name,

		mu:       &sync.Mutex{},
		current:  codec,
		previous: codec,
	}

	t := time.NewTicker(options.Expiry * 9 / 10)
	go func(
		ctx context.Context,
		enc *cookieEncoder,
		dec *cookieDecoder,
		rotate <-chan time.Time,
	) {
		for {
			select {
			case <-ctx.Done():
				return
			case _ = <-rotate: // t := <-rotate to capture rotation time
				fresh := newCodec()
				// TODO: update KV store with fresh codec.

				dec.mu.Lock()
				dec.previous = dec.current
				dec.current = fresh
				dec.mu.Unlock()

				enc.mu.Lock()
				dec.current = fresh
				enc.mu.Unlock()
			}
		}
	}(options.RotationContext, enc, dec, t.C)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(
					&sessionContext{
						Context: r.Context(),
						encoder: enc,
						decoder: dec,

						mu: &sync.Mutex{},
						w:  w,
						r:  r,
					},
				))
			},
		)
	}, nil
}
