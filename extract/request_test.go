package extract_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/extract"
)

type rawRequest struct {
	Host    string
	Address string
	Agent   string
	Method  string
}

func (r *rawRequest) Validate(ctx context.Context) error {
	if len(r.Host) < 1 {
		return errors.New("host is required")
	}
	if len(r.Address) < 1 {
		return errors.New("address is required")
	}
	if len(r.Agent) < 1 {
		return errors.New("agent is required")
	}
	if len(r.Method) < 1 {
		return errors.New("method is required")
	}
	return nil
}

func TestRawRequestValues(t *testing.T) {
	hostExtractor, err := extract.NewHostExtractor()
	if err != nil {
		t.Fatal(err)
	}
	addresExtractor, err := extract.NewRemoteAddressExtractor("address")
	if err != nil {
		t.Fatal(err)
	}
	agentExtractor, err := extract.NewUserAgentExtractor("agent")
	if err != nil {
		t.Fatal(err)
	}
	methodExtractor, err := extract.NewMethodExtractor("method")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/test/raw", htadaptor.Must(
		htadaptor.NewVoidFuncAdaptor(
			func(ctx context.Context, r *rawRequest) error {
				return nil
			},
			htadaptor.WithExtractors(
				hostExtractor,
				addresExtractor,
				agentExtractor,
				methodExtractor,
			),
		),
	))
	req := httptest.NewRequest(http.MethodGet, "/test/raw", nil)
	req.Header.Set("User-Agent", "Testing User Agent")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("status code:", res.StatusCode)
		t.Log("body:", string(data))
		t.Fatal("request failed")
	}
	// if err != nil {
	// 	t.Errorf("expected error to be nil got %v", err)
	// }
	//
	// var response testResponse
	// t.Logf("response: %s", data)
	// err = json.NewDecoder(bytes.NewReader(data)).Decode(&response)
}
