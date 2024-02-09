package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

func TestCookieValue(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodGet, "/test/cookie", nil)
	req.AddCookie(&http.Cookie{
		Name:  "UUID",
		Value: "testUUID",
	})

	validateHandler(t, mux, []*testCase{
		{
			Request:  req,
			Response: testResponse{Value: "testUUID"},
			Failure:  "extracted cookie value does not match",
		},
	})
}
