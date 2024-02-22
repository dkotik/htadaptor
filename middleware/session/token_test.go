package session

import (
	"testing"
	"time"
)

func TestUTC(t *testing.T) {
	tm := time.Now()
	if tm.UTC().Unix() != tm.Unix() {
		t.Fatal("UTC unix int64 value does not match local")
	}
}

func TestHMAC(t *testing.T) {
	k, err := NewKey(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	signed := k.Sign([]byte("test message"))
	if !k.Validate(signed) {
		t.Fatalf("signature does not match: %s", signed)
	}
}
