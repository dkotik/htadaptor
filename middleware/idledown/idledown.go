/*
Package idledown provides an [htadaptor.Middleware] that shuts
the server down due to inactivity.
*/
package idledown

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dkotik/htadaptor"
)

// New creates an [htadaptor.Middleware] that resets its internal timer
// for each [http.Request] and shuts down
// an [http.Server] when the timer runs out. Use it for services
// that are meant to scale down to zero or few instances.
func New(
	parent context.Context,
	shutdownAfter time.Duration,
) (context.Context, htadaptor.Middleware) {
	if shutdownAfter < time.Second {
		panic("idle duration must be greater than zero")
	}
	if parent == nil {
		panic("cannot use a <nil> context")
	}
	ctx, cancel := context.WithCancelCause(parent)
	timer := time.NewTimer(shutdownAfter)
	go func(ctx context.Context, waitingOn <-chan time.Time, d time.Duration) {
		select {
		case <-ctx.Done():
			// nothing
		case <-waitingOn:
			cancel(fmt.Errorf("HTTP handlers were idle for more than %.2f minutes", float32(d)/float32(time.Minute)))
		}
	}(ctx, timer.C, shutdownAfter)

	return ctx, func(next http.Handler) http.Handler {
		if next == nil {
			panic("cannot use a <nil> next handler")
		}
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// OPTIMIZE: documentation seems unclear if Reset is concurrency safe?
				timer.Reset(shutdownAfter)
				// if ResponseWriter is a reverse proxy, see:
				//  - https://github.com/superfly/tired-proxy
				//  - https://github.com/ties-v/tired-proxy
				// r.Host = remote.Host
				next.ServeHTTP(w, r)
			},
		)
	}
}
