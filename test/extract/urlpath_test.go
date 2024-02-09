package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

func TestURLPathValue(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/test/{UUID}", htadaptor.Must(
		htadaptor.NewUnaryFuncAdaptor(
			func(ctx context.Context, r *testRequest) (*testResponse, error) {
				return &testResponse{
					Value: r.UUID,
				}, nil
			},
			htadaptor.WithPathValues("UUID"),
		),
	))
	validateHandler(t, mux, []*testCase{
		{
			Request:  httptest.NewRequest(http.MethodGet, "/test/123", nil),
			Response: testResponse{Value: "123"},
			Failure:  "extracted cookie value does not match",
		},
	})
}
