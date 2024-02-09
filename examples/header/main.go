/*
Package main demonstrates the use of header extractor for request decoding.
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

func (t *testRequest) Validate() error {
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
	mux.Handle("/test/header", htadaptor.Must(
		htadaptor.NewUnaryFuncAdaptor(
			func(ctx context.Context, r *testRequest) (*testResponse, error) {
				return &testResponse{
					Value: r.UUID,
				}, nil
			},
			htadaptor.WithHeaderValues("UUID"),
		),
	))

	fmt.Printf(
		`Listening at http://%[1]s/

    Test Header Value Extraction:
      curl -v -H "UUID: testUUID" http://%[1]s/test/header
`,
		l.Addr(),
	)

	http.Serve(l, mux)
}
