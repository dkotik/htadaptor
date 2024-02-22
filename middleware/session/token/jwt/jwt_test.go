package jwt

import (
	"reflect"
	"testing"
)

func TestDecoing(t *testing.T) {
	tokens := New()
	// if err != nil {
	// 	t.Fatal(err)
	// }

	data := map[string]any{
		"one": float64(1.0),
		"two": "two",
	}
	token, err := tokens.Encode(data)
	if err != nil {
		t.Fatal(err)
	}
	var another map[string]any
	err = tokens.Decode(&another, token)
	if err != nil {
		t.Fatal(err)
	}

	// for k, v := range data {
	// 	if !reflect.DeepEqual(another[k], v) {
	// 		t.Logf("original: %+v", v)
	// 		t.Logf("decoded: %+v", another[k])
	// 		t.Fatal("decoded value does not match original")
	// 	}
	// }

	if !reflect.DeepEqual(data, another) {
		t.Logf("original: %+v", data)
		t.Logf("decoded: %+v", another)
		t.Fatal("decoded value does not match original")
	}
}
