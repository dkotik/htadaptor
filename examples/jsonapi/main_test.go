package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

/*
 *     Test Order (Unary):
   curl -v -d '{"item":"box","quantity":1}' -H 'Content-Type: application/json' http://%[1]s/api/v1/order/1
 Test Price (Unary String):
   curl -v -G -d 'item=shirt' http://%[1]s/api/v1/price
 Test Inventory (Nullary):
   curl -v http://%[1]s/api/v1/inventory
 Test Record (Unary Void):
   curl -v -d '{"item":"box","quantity":1}' -H 'Content-Type: application/json' http://%[1]s/api/v1/record
*/

func TestCookieExample(t *testing.T) {
	mux := newService(&OnlineStore{})

	srv := httptest.NewTestServer(t, mux)
	client := srv.Client()
	defer srv.Close()

	doRequest := func(req *http.Request) ([]byte, error) {
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}

	// /api/v1/order/{number}
	formData := url.Values{}
	formData.Set("item", "shirt")
	formData.Set("quantity", "1")
	payload := strings.NewReader(formData.Encode())
	req, err := http.NewRequest(
		"POST",
		srv.URL+strings.Replace(pathOrder, "{number}", "1", 1),
		payload,
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := doRequest(req)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	if !bytes.Equal(body, []byte("true\n")) {
		t.Errorf("Unexpected body: %s", string(body))
	}

	// pathPrice     = "/api/v1/price"
	req, err = http.NewRequest(
		"GET",
		srv.URL+pathPrice+"?item=shirt",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	body, err = doRequest(req)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	if !bytes.Equal(body, []byte("10.99\n")) {
		t.Errorf("Unexpected body: %s", string(body))
	}

	// pathInventory = "/api/v1/inventory"
	req, err = http.NewRequest(
		"GET",
		srv.URL+pathInventory,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	body, err = doRequest(req)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	if !bytes.Equal(body, []byte("[\"shirt\",\"pants\",\"hat\"]\n")) {
		t.Errorf("Unexpected body: %s", string(body))
	}
}
