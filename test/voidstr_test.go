package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/extract"
)

var voidStringCases = []TestCaseJSON[testResponse]{
	{
		Name:     "simple void string request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/voidstr", nil),
		Response: nil,
	},
}

// var voidStringErrorCases = []TestCaseJSON[errorResponse]{
// 	{
// 		Name:     "empty void string request",
// 		Request:  httptest.NewRequest(http.MethodGet, "/test/voidstr", nil),
// 		Response: &errorResponse{Error: "invalid request: UUID is empty"},
// 	},
// }

func TestVoidStringRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/test/voidstr", htadaptor.Must(
		htadaptor.NewVoidStringFuncAdaptor(
			func(ctx context.Context, s string) error {
				return nil
			},
			extract.StringValueExtractorFunc(
				func(r *http.Request) (string, error) {
					return "test string", nil
				},
			),
		),
	))

	TestJSON(t, mux, voidStringCases)
	// TestJSON(t, mux, voidStringErrorCases)
}
