package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieExample(t *testing.T) {
	handler, err := newHandlerHTMX()
	if err != nil {
		t.Fatalf("Failed to create cookie handler: %v", err)
	}

	srv := httptest.NewTestServer(t, handler)
	client := srv.Client()
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL+handlerPath, nil)
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
	b := &bytes.Buffer{}
	err = greetingTemplate.Execute(b, testResponse{
		Name: "Guest",
	})
	if err != nil {
		t.Fatalf("Failed to marshal expected body: %v", err)
	}
	if !bytes.Equal(body, b.Bytes()) {
		t.Errorf("Unexpected body: %s", string(body))
	}
}
