/*
Package main demonstrates the simplest implementation of a rotating
session context.
*/
package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/dkotik/htadaptor/middleware/session"
)

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	sessionMiddleware, err := session.New()
	if err != nil {
		panic(err)
	}
	handler := sessionMiddleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			sessionContext, sessionUnlock := session.Lock(r.Context())
			defer sessionUnlock()

			previous, _ := sessionContext.Get("key").(int64)
			current := time.Now().Unix()
			sessionContext.Set("key", current)
			if err := sessionContext.Commit(); err != nil {
				panic(err)
			}

			fmt.Fprintf(w, "previous: %d; current: %d; id: %s", previous, current, sessionContext.ID())
		},
	))

	fmt.Printf(
		`Listening at http://%[1]s/

    Test Session Assignment:
      curl -v http://%[1]s/
`,
		l.Addr(),
	)

	http.Serve(l, handler)
}
