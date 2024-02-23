/*
Package main demonstrates the use of cookie extractor for request decoding.
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

type testRequest struct {
	UUID string
}

func (t *testRequest) Validate(ctx context.Context) error {
	if t.UUID == "" {
		return errors.New("UUID is empty")
	}
	return nil
}

type testResponse struct {
	Value string
}

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	mux := http.NewServeMux()
	mux.Handle("/test/cookie", htadaptor.Must(
		htadaptor.NewUnaryFuncAdaptor(
			func(ctx context.Context, r *testRequest) (*testResponse, error) {
				return &testResponse{
					Value: r.UUID,
				}, nil
			},
			htadaptor.WithCookieValues("UUID"),
		),
	))

	fmt.Printf(
		`Listening at http://%[1]s/

    Test Cookie Value Extraction:
      curl -v -b "UUID=testUUID" http://%[1]s/test/cookie
`,
		l.Addr(),
	)

	http.Serve(l, mux)
}
