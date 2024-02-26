package extract_test

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
				if r.SomeValue != "test value" {
					t.Fatal("request header with dash failed to decode in request")
				}
				return &testResponse{
					Value: r.UUID,
				}, nil
			},
			htadaptor.WithHeaderValues("UUID", "Some-Value"),
		),
	))
	req := httptest.NewRequest(http.MethodGet, "/test/header", nil)
	req.Header.Add("UUID", "testUUID")
	req.Header.Add("Some-Value", "test value")

	validateHandler(t, mux, []*testCase{
		{
			Request:  req,
			Response: testResponse{Value: "testUUID"},
			Failure:  "extracted cookie value does not match",
		},
	})
}
