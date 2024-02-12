package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/extract"
)

var unaryStringCases = []TestCaseJSON[testResponse]{
	{
		Name:     "simple unary string request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/unarystr", nil),
		Response: &testResponse{Value: "test string"},
	},
}

// var unaryStringErrorCases = []TestCaseJSON[errorResponse]{
// 	{
// 		Name:     "empty unary string request",
// 		Request:  httptest.NewRequest(http.MethodGet, "/test/unarystr", nil),
// 		Response: &errorResponse{Error: "invalid request: UUID is empty"},
// 	},
// }

func TestUnaryStringRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/test/unarystr", htadaptor.Must(
		htadaptor.NewUnaryStringFuncAdaptor(
			func(ctx context.Context, s string) (*testResponse, error) {
				return &testResponse{
					Value: s,
				}, nil
			},
			extract.StringValueExtractorFunc(
				func(r *http.Request) (string, error) {
					return "test string", nil
				},
			),
		),
	))

	TestJSON(t, mux, unaryStringCases)
	// TestJSON(t, mux, unaryStringErrorCases)
}
