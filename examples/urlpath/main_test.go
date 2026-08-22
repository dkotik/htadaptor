package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCookieExample(t *testing.T) {
	handler, err := newPathHandler()
	if err != nil {
		t.Fatalf("Failed to create cookie handler: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle(
		handlerPath,
		handler,
	)

	srv := httptest.NewTestServer(t, mux)
	client := srv.Client()
	defer srv.Close()

	const testPath = "xyz123456"
	req, err := http.NewRequest(
		"GET",
		srv.URL+(strings.Replace(handlerPath, "{UUID}", testPath, 1)),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	expected, err := json.Marshal(testResponse{
		Value: testPath,
	})
	if err != nil {
		t.Fatalf("Failed to marshal expected body: %v", err)
	}
	if !bytes.Equal(body, append(expected, '\n')) {
		t.Errorf("Unexpected body: %s", string(body))
	}
}
