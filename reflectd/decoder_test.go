package reflectd_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dkotik/htadaptor/reflectd"
)

type testRequest struct {
	TestField    string
	AnotherField string
	TestHeader   string
}

func (t testRequest) ToRequestHTTP() *http.Request {
	formData := url.Values{}
	formData.Set("TestField", t.TestField)
	formData.Set("AnotherField", t.AnotherField)
	formData.Set("TestHeader", t.TestHeader)
	payload := strings.NewReader(formData.Encode())
	req := httptest.NewRequest(http.MethodPost, "/login", payload)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestDecoder(t *testing.T) {
	d, err := reflectd.NewDecoder()
	if err != nil {
		t.Fatal(err)
	}

	original := testRequest{TestField: "field", AnotherField: "another", TestHeader: "header"}
	v := testRequest{}
	if err := d.Decode(&v, original.ToRequestHTTP()); err != nil {
		t.Fatal(err)
	}
	if v.TestField != original.TestField {
		t.Errorf("expected TestField to be 'field', got %q", v.TestField)
	}
	if v.AnotherField != original.AnotherField {
		t.Errorf("expected AnotherField to be 'another', got %q", v.AnotherField)
	}
	if v.TestHeader != original.TestHeader {
		t.Errorf("expected TestHeader to be 'header', got %q", v.TestHeader)
	}

	p := asPointer[testRequest, *testRequest]()
	if err := d.Decode(p, original.ToRequestHTTP()); err != nil {
		t.Fatal(err)
	}
	if p.TestField != original.TestField {
		t.Errorf("expected TestField to be 'field', got %q", p.TestField)
	}
	if p.AnotherField != original.AnotherField {
		t.Errorf("expected AnotherField to be 'another', got %q", p.AnotherField)
	}
	if p.TestHeader != original.TestHeader {
		t.Errorf("expected TestHeader to be 'header', got %q", p.TestHeader)
	}
}

func asPointer[T any, P *T]() *T {
	ptr := new(T)
	return P(ptr)
}
