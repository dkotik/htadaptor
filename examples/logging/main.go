/*
Package main demonstrates the use of custom logger with an adaptor.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/dkotik/htadaptor"
)

type testResponse struct {
	Name string
}

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	mux := http.NewServeMux()
	mux.Handle("/test/logger", htadaptor.Must(
		htadaptor.NewNullaryFuncAdaptor(
			func(ctx context.Context) (*testResponse, error) {
				return nil, errors.New("logging error")
			},
			htadaptor.WithLogger(htadaptor.LoggerFunc(
				func(r *http.Request, err error) {
					fmt.Printf("logging request error state: %+v\n", err)
				},
			)),
		),
	))

	fmt.Printf(
		`Listening at http://%[1]s/

    Test custom Logger:
      curl -v http://%[1]s/test/logger
`,
		l.Addr(),
	)

	http.Serve(l, mux)
}
