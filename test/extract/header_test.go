package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

func TestHeaderValue(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodGet, "/test/header", nil)
	req.Header.Add("UUID", "testUUID")

	validateHandler(t, mux, []*testCase{
		{
			Request:  req,
			Response: testResponse{Value: "testUUID"},
			Failure:  "extracted cookie value does not match",
		},
	})
}
