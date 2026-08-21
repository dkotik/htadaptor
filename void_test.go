package htadaptor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

var voidCases = []testCaseJSON[testResponse]{
	{
		Name: "simple void request",
		Request: NewPostRequestJSON("/test/void", &testRequest{
			UUID: "testUUID",
		}),
		Response: nil,
	},
}

var voidErrorCases = []testCaseJSON[errorResponse]{
	{
		Name:     "empty void request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/void", nil),
		Response: &errorResponse{Error: "UUID is empty"},
	},
}

func TestVoidRequest(t *testing.T) {
	mux := http.NewServeMux()
	adaptor, err := htadaptor.New()
	if err != nil {
		t.Fatal(err)
	}
	mux.Handle("/test/void",
		adaptor.AdaptVoidFunc(
			func(ctx context.Context, r *testRequest) error {
				return nil
			},
		),
	)

	runCasesJSON(t, mux, voidCases)
	runCasesJSON(t, mux, voidErrorCases)
}
