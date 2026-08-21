package htadaptor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

var unaryCases = []testCaseJSON[testResponse]{
	{
		Name: "simple unary request",
		Request: NewPostRequestJSON("/test/unary", &testRequest{
			UUID: "testUUID",
		}),
		Response: &testResponse{Value: "testUUID"},
	},
}

var unaryErrorCases = []testCaseJSON[errorResponse]{
	{
		Name:     "empty unary request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/unary", nil),
		Response: &errorResponse{Error: "UUID is empty"},
	},
}

func TestUnaryRequest(t *testing.T) {
	// t.Skip("decoder behaves strangely")
	mux := http.NewServeMux()
	adaptor, err := htadaptor.New()
	if err != nil {
		t.Fatal(err)
	}
	mux.Handle("/test/unary",
		adaptor.AdaptFunc(
			func(ctx context.Context, r *testRequest) (testResponse, error) {
				return testResponse{
					Value: r.UUID,
				}, nil
			},
		),
	)
	// mux.Handle("/test/unary/value",
	// 	adaptor.AdaptFunc(
	// 		func(ctx context.Context, r testRequest) (testResponse, error) {
	// 			return testResponse{
	// 				Value: r.UUID,
	// 			}, nil
	// 		},
	// 	),
	// )

	runCasesJSON(t, mux, unaryCases)
	runCasesJSON(t, mux, unaryErrorCases)
}
