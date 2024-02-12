package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

var voidCases = []TestCaseJSON[testResponse]{
	{
		Name: "simple void request",
		Request: NewPostRequestJSON("/test/void", &testRequest{
			UUID: "testUUID",
		}),
		Response: nil,
	},
}

var voidErrorCases = []TestCaseJSON[errorResponse]{
	{
		Name:     "empty void request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/void", nil),
		Response: &errorResponse{Error: "invalid request: UUID is empty"},
	},
}

func TestVoidRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/test/void", htadaptor.Must(
		htadaptor.NewVoidFuncAdaptor(
			func(ctx context.Context, r *testRequest) error {
				return nil
			},
		),
	))

	TestJSON(t, mux, voidCases)
	TestJSON(t, mux, voidErrorCases)
}
