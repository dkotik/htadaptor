/*
Package test ensures that all htadaptor components work.
*/
package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

type testRequest struct {
	UUID string
}

func (t *testRequest) Validate() error {
	if t.UUID == "" {
		return errors.New("UUID is empty")
	}
	return nil
}

type testResponse struct {
	Value string
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
