/*
Package test provides tooling and routines for building and running htadaptor tests.
*/
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type TestCaseJSON[T any] struct {
	Name     string
	Request  *http.Request
	Response *T
}

func TestJSON[T any](t *testing.T, h http.Handler, cases []TestCaseJSON[T]) {
	for _, tc := range cases {
		t.Run(tc.Name,
			func(t *testing.T) {
				data, code, header := CaptureResponse(h, tc.Request)
				if len(data) < 1 && tc.Response == nil {
					if code != http.StatusNoContent {
						t.Fatal("got no data, but status code does not match", code, http.StatusText(code))
					}
					return // got nothin, while expecting nil
				}
				if header.Get("content-type") != "application/json" {
					t.Fatal("expected a JSON response, instead got:", header.Get("content-type"))
				}
				t.Logf("response: %s", data)

				var response *T
				err := json.NewDecoder(bytes.NewReader(data)).Decode(&response)
				if err != nil {
					t.Fatal("failed to decode JSON:", err.Error())
				}
				if !reflect.DeepEqual(tc.Response, response) {
					expected, err := json.Marshal(tc.Response)
					if err != nil {
						panic(err)
					}
					t.Fatalf("expected: %s", expected)
				}
			},
		)
	}
}

func CaptureResponse(h http.Handler, r *http.Request) ([]byte, int, http.Header) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(fmt.Errorf("unable to copy data: %w", err))
	}
	return data, res.StatusCode, res.Header
}

func NewPostRequestJSON(p string, v any) (r *http.Request) {
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(v); err != nil {
		panic(err)
	}
	r = httptest.NewRequest(http.MethodPost, p, b)
	r.Header.Set("content-type", "application/json")
	return r
}

type testRequest struct {
	UUID string
}

func (t *testRequest) Validate(ctx context.Context) error {
	if t.UUID == "" {
		return errors.New("UUID is empty")
	}
	return nil
}

type testResponse struct {
	Value string
}

type errorResponse struct {
	Error string `json:"error"`
}
