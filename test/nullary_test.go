package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

var nullaryCases = []TestCaseJSON[testResponse]{
	{
		Name:     "simple nullary request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/nullary", nil),
		Response: &testResponse{Value: "testUUID"},
	},
}

func TestNullaryRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/test/nullary", htadaptor.Must(
		htadaptor.NewNullaryFuncAdaptor(
			func(ctx context.Context) (*testResponse, error) {
				return &testResponse{
					Value: "testUUID",
				}, nil
			},
		),
	))

	TestJSON(t, mux, nullaryCases)
}
