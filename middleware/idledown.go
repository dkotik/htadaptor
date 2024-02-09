package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// IdleDown resets its timer for each [http.Request] and shuts down
// an [http.Server] when the timer runs out. Use it for services
// that are meant to scale down to zero or few instances.
type IdleDown struct {
	idle  time.Duration
	timer *time.Timer
	next  http.Handler
}

func NewIdleDown(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		idleDown := &IdleDown{
			idle:  d,
			timer: time.NewTimer(d),
			next:  next,
		}

		go func(waitingOn <-chan time.Time, d time.Duration) {
			<-waitingOn
			fmt.Printf("Shutting down service because there were no HTTP requests for %.2f minutes.\n", float32(d)/float32(time.Minute))
			os.Exit(0)
		}(idleDown.timer.C, d)

		return idleDown
	}
}

func (d *IdleDown) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// OPTIMIZE: documentation seems unclear if Reset is concurrency safe?
	d.timer.Reset(d.idle)
	// if ResponseWriter is a reverse proxy, see:
	//  - https://github.com/superfly/tired-proxy
	//  - https://github.com/ties-v/tired-proxy
	// r.Host = remote.Host
	d.next.ServeHTTP(w, r)
}
