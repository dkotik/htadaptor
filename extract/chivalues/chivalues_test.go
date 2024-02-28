package chivalues

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
)

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

func TestChiValues(t *testing.T) {
	extractor, err := New("UUID")
	if err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Get("/test/{UUID}", func(w http.ResponseWriter, r *http.Request) {
		values := url.Values{}
		if err := extractor.ExtractRequestValue(values, r); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(&testResponse{
			Value: values.Get("UUID"),
		})
		if err != nil {
			t.Fatal(err)
		}
		w.Write(data)
	})

	validateHandler(t, r, []*testCase{
		{
			Request:  httptest.NewRequest(http.MethodGet, "/test/123", nil),
			Response: testResponse{Value: "123"},
			Failure:  "extracted cookie value does not match",
		},
	})
}

type testCase struct {
	Request  *http.Request
	Response testResponse
	Failure  string
}

func validateHandler(t *testing.T, h http.Handler, cases []*testCase) {
	for i, tc := range cases {
		t.Run("validating handler - test case #"+strconv.Itoa(i),
			func(t *testing.T) {
				validateResponse(t, h, tc)
			},
		)
	}
}

func validateResponse(t *testing.T, h http.Handler, c *testCase) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, c.Request)
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	var response testResponse
	err = json.NewDecoder(bytes.NewReader(data)).Decode(&response)
	if err != nil {
		t.Fatal("failed to decode JSON:", err.Error())
	}

	if !reflect.DeepEqual(&c.Response, &response) {
		t.Logf("response: %s", data)
		expected, err := json.Marshal(c.Response)
		if err != nil {
			panic(err)
		}
		t.Logf("expected: %s", expected)
		t.Fatalf("test case failed: %s", c.Failure)
	}
}
