package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieExample(t *testing.T) {
	handler, err := newCookieHandler()
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
	cookie := &http.Cookie{
		Name:  "UUID",
		Value: "xyz123456",
	}
	req.AddCookie(cookie)

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
		Value: cookie.Value,
	})
	if err != nil {
		t.Fatalf("Failed to marshal expected body: %v", err)
	}
	if !bytes.Equal(body, append(expected, '\n')) {
		t.Errorf("Unexpected body: %s", string(body))
	}
}
