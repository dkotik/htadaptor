/*
Package main demonstrates the use of URL path value extractor for request decoding.
*/
package main

import (
	"context"
	"errors"
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

func newPathHandler() (http.Handler, error) {
	adaptor := htadaptor.New()
	return adaptor.AdaptFunc(
		func(ctx context.Context, r *testRequest) (*testResponse, error) {
			return &testResponse{
				Value: r.UUID,
			}, nil
		},
		htadaptor.WithPathValues("UUID"),
	)
}

const handlerPath = "/test/{UUID}"

func main() {
	mux := http.NewServeMux()
	mux.Handle(
		handlerPath,
		htadaptor.Must(newPathHandler()),
	)
	http.ListenAndServe(":0", mux)
}
