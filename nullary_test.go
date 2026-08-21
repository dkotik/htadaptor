package htadaptor_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
)

var nullaryCases = []testCaseJSON[testResponse]{
	{
		Name:     "simple nullary request",
		Request:  httptest.NewRequest(http.MethodGet, "/test/nullary", nil),
		Response: &testResponse{Value: "testUUID"},
	},
}

func TestNullaryRequest(t *testing.T) {
	mux := http.NewServeMux()
	adaptor, err := htadaptor.New()
	if err != nil {
		t.Fatal(err)
	}
	mux.Handle("/test/nullary",
		adaptor.AdaptNullaryFunc(
			func(ctx context.Context) (*testResponse, error) {
				return &testResponse{
					Value: "testUUID",
				}, nil
			},
		),
	)

	runCasesJSON(t, mux, nullaryCases)
}
