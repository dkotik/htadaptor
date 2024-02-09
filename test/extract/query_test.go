package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

func TestQueryValue(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/test/query", htadaptor.Must(
		htadaptor.NewUnaryFuncAdaptor(
			func(ctx context.Context, r *testRequest) (*testResponse, error) {
				return &testResponse{
					Value: r.UUID,
				}, nil
			},
			htadaptor.WithQueryValues("UUID"),
		),
	))
	validateHandler(t, mux, []*testCase{
		{
			Request:  httptest.NewRequest(http.MethodGet, "/test/query?UUID=fromQuery", nil),
			Response: testResponse{Value: "fromQuery"},
			Failure:  "extracted cookie value does not match",
		},
	})
}
