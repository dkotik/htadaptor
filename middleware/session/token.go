package session

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"slices"
	"sync"

	"github.com/dkotik/htadaptor/middleware/session/secrets"
)

const (
	tokenSignatureSize = 64
	tokenLabelSize     = 6
	tokenTagSize       = tokenSignatureSize + tokenLabelSize
)

type Tokenizer interface {
	Encode(any) (string, error)
	Decode(any, string) error
}

type hmacTokenizer struct {
	wmu   *sync.Mutex
	write *secrets.Secret

	rmu     *sync.Mutex
	present *secrets.Secret
	past    *secrets.Secret
}

func NewTokenizer(withOptions ...secrets.Option) Tokenizer {
	t := &hmacTokenizer{
		wmu: &sync.Mutex{},
		rmu: &sync.Mutex{},
	}

	if err := secrets.NewRotation(t.Rotate, withOptions...); err != nil {
		panic(err) // TODO: beautify.
	}
	return t
}

func (h *hmacTokenizer) Encode(data any) (string, error) {
	b := &bytes.Buffer{}
	if err := gob.NewEncoder(b).Encode(data); err != nil {
		return "", err
	}
	encoded := make([]byte, base64.RawURLEncoding.EncodedLen(b.Len()))
	base64.RawURLEncoding.Encode(encoded, b.Bytes())

	h.wmu.Lock()
	defer h.wmu.Unlock()
	return string(Sign(h.write, encoded)), nil
}

// Validate return true if given bytes begin with hex-encoded
// signature that matches HMAC sum of [Key.Label] followed
// by the rest of the data.
func Validate(secret *secrets.Secret, b []byte) bool {
	// if len(b) <= tokenTagSize {
	// 	return false
	// }
	// if !bytes.Equal(k.Label, b[:12]) {
	// 	// does not match key label
	// 	return false
	// }
	// sha256 block size is 64, but signature length is 32 bytes
	mac := hmac.New(sha256.New, secret.Entropy)
	_, err := mac.Write(b[tokenSignatureSize:])
	if err != nil {
		return false
	}
	signature := make([]byte, 64)
	_ = hex.Encode(signature, mac.Sum(nil)) // H(key ∥ H(key ∥ message))
	return hmac.Equal(b[:tokenSignatureSize], signature)
}

func Sign(secret *secrets.Secret, b []byte) []byte {
	mac := hmac.New(sha256.New, secret.Entropy)
	_, err := mac.Write(secret.ID)
	if err != nil {
		return nil
	}
	_, err = mac.Write(b)
	if err != nil {
		return nil
	}
	signature := make([]byte, tokenSignatureSize) // 32 x 2 for hex
	_ = hex.Encode(signature, mac.Sum(nil))       // H(key ∥ H(key ∥ message))
	// panic(string(signature))
	return slices.Concat(signature, secret.ID, b)
}

func (h *hmacTokenizer) Decode(data any, token string) (err error) {
	b := []byte(token)
	h.rmu.Lock()
	if !Validate(h.present, b) && !Validate(h.past, b) {
		// TODO: // OPTIMIZE validation
		return nil
	}
	h.rmu.Unlock()
	b = b[tokenTagSize:] // chomp off tag
	dbuf := make([]byte, base64.RawURLEncoding.DecodedLen(len(b)))
	if _, err = base64.RawURLEncoding.Decode(dbuf, b); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewReader(dbuf)).Decode(data)
}

func (h *hmacTokenizer) Rotate(present, past *secrets.Secret) error {
	h.rmu.Lock()
	h.present = present
	h.past = past
	h.rmu.Unlock()

	h.wmu.Lock()
	h.write = present
	h.wmu.Unlock()
	return nil
}
