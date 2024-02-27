package htadaptor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/extract"
)

var unaryStringCases = []testCaseJSON[testResponse]{
	{
		Name:     "simple unary string request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/unarystr", nil),
		Response: &testResponse{Value: "test string"},
	},
}

// var unaryStringErrorCases = []testCaseJSON[errorResponse]{
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

	runCasesJSON(t, mux, unaryStringCases)
	// runCasesJSON(t, mux, unaryStringErrorCases)
}
