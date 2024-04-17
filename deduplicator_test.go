package htadaptor_test

import (
	"context"
	"testing"
	"time"

	"github.com/dkotik/htadaptor"
)

func TestDeduplicator(t *testing.T) {
	window := time.Millisecond * 20
	d := htadaptor.NewDeduplicator(context.Background(), window)
	testCase := func(request any, expect bool) {
		ok, err := d.IsDuplicate(request)
		if err != nil {
			t.Error(err)
		}
		if expect != ok {
			t.Errorf("expectation not met: %v vs %v", expect, ok)
		}
	}

	testCase("", false)
	testCase("", true)

	basic := &testRequest{UUID: "test"}
	testCase(basic, false)
	testCase(basic, true)
	testCase(basic, true)

	complex := []*testRequest{
		basic,
		basic,
		basic,
	}
	testCase(complex, false)
	testCase(complex, true)
	testCase(complex, true)
	complex[1].UUID = "nothing"
	testCase(complex, false)
	testCase(complex, true)

	if total := d.Len(); total != 4 {
		t.Errorf("unexpected number of tracked records: %d vs %d", 6, total)
	}

	<-time.After(window * 2)
	if total := d.Len(); total != 0 {
		t.Errorf("clean up procedure failed: %d records are still being tracked", total)
	}
}
