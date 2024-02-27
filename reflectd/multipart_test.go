package reflectd_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/dkotik/htadaptor/reflectd"
)

func TestMultipart(t *testing.T) {
	decoder, err := reflectd.NewDecoder(
		reflectd.WithQueryValues("anotherField"),
		reflectd.WithHeaderValues("testHeader"),
	)
	if err != nil {
		t.Fatal(err)
	}

	var mp bytes.Buffer
	w := multipart.NewWriter(&mp)
	field, err := w.CreateFormField("testField")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = field.Write([]byte("testValue")); err != nil {
		t.Fatal(err)
	}

	f, err := w.CreateFormFile("random", "random.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write([]byte("randomContent")); err != nil {
		t.Fatal(err)
	}

	if err = w.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "somepath?anotherField=anotherValue", bytes.NewReader(mp.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	// panic(w.FormDataContentType())
	ct := w.FormDataContentType()
	// ct = strings.Replace(ct, "multipart/form-data", "multipart/mixed", 1)
	req.Header.Add("Content-Type", ct)
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
