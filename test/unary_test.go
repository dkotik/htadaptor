package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

var unaryCases = []TestCaseJSON[testResponse]{
	{
		Name: "simple unary request",
		Request: NewPostRequestJSON("/test/unary", &testRequest{
			UUID: "testUUID",
		}),
		Response: &testResponse{Value: "testUUID"},
	},
}

var unaryErrorCases = []TestCaseJSON[errorResponse]{
	{
		Name:     "empty unary request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/unary", nil),
		Response: &errorResponse{Error: "UUID is empty"},
	},
}

func TestUnaryRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/test/unary", htadaptor.Must(
		htadaptor.NewUnaryFuncAdaptor(
			func(ctx context.Context, r *testRequest) (*testResponse, error) {
				return &testResponse{
					Value: r.UUID,
				}, nil
			},
		),
	))

	TestJSON(t, mux, unaryCases)
	TestJSON(t, mux, unaryErrorCases)
}
