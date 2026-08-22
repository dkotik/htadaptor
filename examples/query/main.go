/*
Package main demonstrates the use of URL query value extractor for request decoding.
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

const handlerPath = "/test/query"

func newQueryHandler() (http.Handler, error) {
	adaptor := htadaptor.New()
	return adaptor.AdaptFunc(
		func(ctx context.Context, r *testRequest) (*testResponse, error) {
			return &testResponse{
				Value: r.UUID,
			}, nil
		},
		htadaptor.WithQueryValues("UUID"),
	)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle(
		handlerPath,
		htadaptor.Must(newQueryHandler()),
	)
	http.ListenAndServe(":0", mux)
}
