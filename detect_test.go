package htadaptor_test

import (
	"context"
	"testing"

	"github.com/dkotik/htadaptor"
)

func TestFunctionTypeAssertion(t *testing.T) {
	nullary := func(_ context.Context) (int, error) {
		return 0, nil
	}
	detected, err := htadaptor.Detect(nullary)
	if err != nil {
		t.Fatal(err)
	}
	if detected != htadaptor.FuncTypeNullary {
		t.Fatalf("type detection failed: expected %q, but got %q", htadaptor.FuncTypeNullary, detected)
	}

	unary := func(_ context.Context, _ *testRequest) (int, error) {
		return 0, nil
	}
	detected, err = htadaptor.Detect(unary)
	if err != nil {
		t.Fatal(err)
	}
	if detected != htadaptor.FuncTypeUnary {
		t.Fatalf("type detection failed: expected %q, but got %q", htadaptor.FuncTypeUnary, detected)
	}

	void := func(_ context.Context, _ *testRequest) error {
		return nil
	}
	detected, err = htadaptor.Detect(void)
	if err != nil {
		t.Fatal(err)
	}
	if detected != htadaptor.FuncTypeVoid {
		t.Fatalf("type detection failed: expected %q, but got %q", htadaptor.FuncTypeVoid, detected)
	}
}
