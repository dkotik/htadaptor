package decoder

import (
	"net/http"
	"strings"
	"testing"
)

type testRequest struct {
	TestField    string
	AnotherField string
	TestHeader   string
}

func TestDecoder(t *testing.T) {
	decoder, err := New(
		WithQueryValues("anotherField"),
		WithHeaderValues("testHeader"),
	)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "somepath?anotherField=anotherValue", strings.NewReader(
		`testField=testValue`,
	))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("testHeader", "testHeaderValue")

	v := &testRequest{}
	if err = decoder.Decode(v, req); err != nil {
		t.Fatal(err)
	}
	if v.TestField != `testValue` {
		t.Fatal("failed to decode testField value")
	}
	if v.AnotherField != `anotherValue` {
		t.Fatal("failed to decode another field from URL query")
	}
	if v.TestHeader != "testHeaderValue" {
		t.Fatal("failed to decode HTTP header value")
	}
}
