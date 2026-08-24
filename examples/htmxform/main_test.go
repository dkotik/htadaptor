package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dkotik/htadaptor/middleware/acceptlanguage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sebdah/goldie/v2"
	"golang.org/x/text/language"
)

func TestContactFormExample(t *testing.T) {
	letter := new(Letter)
	handler, err := NewContactForm(&mockSender{Letter: letter})
	if err != nil {
		t.Fatalf("Failed to create contact form: %v", err)
	}
	handler = acceptlanguage.New(i18n.NewBundle(language.English))(handler)

	srv := httptest.NewTestServer(t, handler)
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

	t.Run("GET", func(t *testing.T) {
		req, err := http.NewRequest("GET", srv.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		body, err := doRequest(req)
		if err != nil {
			t.Fatalf("Failed to send GET request: %v", err)
		}
		goldie.New(t).Assert(t, "get", body)
	})

	t.Run("PATCH", func(t *testing.T) {
		b := &bytes.Buffer{}
		formData := url.Values{}
		formData.Set("name", "")
		payload := strings.NewReader(formData.Encode())
		req, err := http.NewRequest("PATCH", srv.URL, payload)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, err := doRequest(req)
		if err != nil {
			t.Fatalf("Failed to send PATCH request: %v", err)
		}
		_, _ = io.Copy(b, bytes.NewReader(body))
		_, _ = b.WriteString("\n\n==================\n\n")

		formData = url.Values{}
		formData.Set("phone", "")
		payload = strings.NewReader(formData.Encode())
		req, err = http.NewRequest("PATCH", srv.URL, payload)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, err = doRequest(req)
		if err != nil {
			t.Fatalf("Failed to send PATCH request: %v", err)
		}
		_, _ = io.Copy(b, bytes.NewReader(body))
		_, _ = b.WriteString("\n\n==================\n\n")

		formData = url.Values{}
		formData.Set("email", "")
		payload = strings.NewReader(formData.Encode())
		req, err = http.NewRequest("PATCH", srv.URL, payload)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, err = doRequest(req)
		if err != nil {
			t.Fatalf("Failed to send PATCH request: %v", err)
		}
		_, _ = io.Copy(b, bytes.NewReader(body))
		_, _ = b.WriteString("\n\n==================\n\n")

		formData = url.Values{}
		formData.Set("message", "")
		payload = strings.NewReader(formData.Encode())
		req, err = http.NewRequest("PATCH", srv.URL, payload)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, err = doRequest(req)
		if err != nil {
			t.Fatalf("Failed to send PATCH request: %v", err)
		}
		_, _ = io.Copy(b, bytes.NewReader(body))

		goldie.New(t).Assert(t, "patch", b.Bytes())
	})

	t.Run("POST", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("name", "")
		formData.Set("phone", "")
		formData.Set("email", "")
		formData.Set("message", "")
		payload := strings.NewReader(formData.Encode())
		req, err := http.NewRequest("POST", srv.URL, payload)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, err := doRequest(req)
		if err != nil {
			t.Fatalf("Failed to send POST request: %v", err)
		}
		goldie.New(t).Assert(t, "post", body)
	})

	t.Run("SENT", func(t *testing.T) {
		mockLetter := Letter{
			Name:    "test name",
			Phone:   "000 000 00000",
			Email:   "test@example.com",
			Message: "test message",
		}
		formData := url.Values{}
		formData.Set("name", mockLetter.Name)
		formData.Set("phone", mockLetter.Phone)
		formData.Set("email", mockLetter.Email)
		formData.Set("message", mockLetter.Message)
		payload := strings.NewReader(formData.Encode())
		req, err := http.NewRequest("POST", srv.URL, payload)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		body, err := doRequest(req)
		if err != nil {
			t.Fatalf("Failed to send POST request: %v", err)
		}
		goldie.New(t).Assert(t, "sent", body)
		if mockLetter.Name != letter.Name {
			t.Errorf("Name mismatch: expected %s, got %s", mockLetter.Name, letter.Name)
		}
		if mockLetter.Phone != letter.Phone {
			t.Errorf("Phone mismatch: expected %s, got %s", mockLetter.Phone, letter.Phone)
		}
		if mockLetter.Email != letter.Email {
			t.Errorf("Email mismatch: expected %s, got %s", mockLetter.Email, letter.Email)
		}
		if mockLetter.Message != letter.Message {
			t.Errorf("Message mismatch: expected %s, got %s", mockLetter.Message, letter.Message)
		}
	})
}
