package htadaptor

import (
	"context"
	"testing"
)

type testRequest struct{}

func (t *testRequest) Validate() error {
	return nil
}

func TestFunctionTypeAssertion(t *testing.T) {
	nullary := func(_ context.Context) (int, error) {
		return 0, nil
	}
	detected, err := Detect(nullary)
	if err != nil {
		t.Fatal(err)
	}
	if detected != FuncTypeNullary {
		t.Fatalf("type detection failed: expected %q, but got %q", FuncTypeNullary, detected)
	}

	unary := func(_ context.Context, _ *testRequest) (int, error) {
		return 0, nil
	}
	detected, err = Detect(unary)
	if err != nil {
		t.Fatal(err)
	}
	if detected != FuncTypeUnary {
		t.Fatalf("type detection failed: expected %q, but got %q", FuncTypeUnary, detected)
	}

	void := func(_ context.Context, _ *testRequest) error {
		return nil
	}
	detected, err = Detect(void)
	if err != nil {
		t.Fatal(err)
	}
	if detected != FuncTypeVoid {
		t.Fatalf("type detection failed: expected %q, but got %q", FuncTypeVoid, detected)
	}
}
