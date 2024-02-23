package jwt

import (
	"fmt"
	"testing"
	"time"
)

func TestDecoing(t *testing.T) {
	tokens := New()
	// if err != nil {
	// 	t.Fatal(err)
	// }

	data := map[string]any{
		"one": float64(1.0),
		"two": "two",
		"exp": time.Now().Add(time.Hour).Unix(),
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

	// if !reflect.DeepEqual(data, another) {
	if fmt.Sprintf("%+v", data) != fmt.Sprintf("%+v", another) {
		t.Logf("original: %+v", data)
		t.Logf("decoded: %+v is <nil>? %v", another, another == nil)
		t.Fatal("decoded value does not match original")
	}
}
