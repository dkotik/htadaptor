/*
Package main demonstrates the simplest implementation of a rotating
session context.
*/
package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/dkotik/htadaptor/middleware/session"
	"github.com/dkotik/htadaptor/middleware/session/token/jwt"
)

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	sessionMiddleware, err := session.New(
		session.WithExpiry(time.Second*5),
		// GorillaKit secure cookie is more tested than the default
		// "github.com/dkotik/htadaptor/middleware/session/token/gorilla"
		// session.WithTokenizer(gorilla.New("test")),

		// Standard JWT is much more common for interop with other systems
		// "github.com/dkotik/htadaptor/middleware/session/token/jwt"
		session.WithTokenizer(jwt.New()),
	)
	if err != nil {
		panic(err)
	}

	logger := slog.New(session.NewSlogHandler(
		slog.NewTextHandler(os.Stderr, nil),
	))

	handler := sessionMiddleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var previous string
			current := fmt.Sprintf("%d", time.Now().Unix())
			var id string

			err := session.Write(r.Context(), func(s session.Session) error {
				previous, _ = s.Get("key").(string)
				s.Set("key", current)
				id = s.ID()
				return nil
			})
			if err != nil {
				panic(err)
			}

			fmt.Fprintf(w, "previous: %s; current: %s; id: %s", previous, current, id)
			logger.InfoContext(
				r.Context(), // important for value injection
				"demonstrating log context injection",
				slog.String("previous", previous),
				slog.String("current", current),
			)
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
