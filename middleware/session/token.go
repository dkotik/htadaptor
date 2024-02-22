package session

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
)

const (
	tokenSignatureSize = 64
	tokenLabelSize     = 6
	tokenTagSize       = tokenSignatureSize + tokenLabelSize
)

type Tokenizer interface {
	Encode(any) (string, error)
	Decode(any, string) error
	Rotate(time.Time) error
}

func NewGorillaSecureCookieTokenizer(name string) Tokenizer {
	sc := securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)
	return &gorillaTokenizer{
		name: name,

		wmu:   &sync.Mutex{},
		write: sc,

		rmu:      &sync.Mutex{},
		current:  sc,
		previous: sc,
	}
}

type gorillaTokenizer struct {
	name string

	wmu   *sync.Mutex
	write *securecookie.SecureCookie

	rmu      *sync.Mutex
	current  *securecookie.SecureCookie
	previous *securecookie.SecureCookie
}

func (g *gorillaTokenizer) Encode(data any) (string, error) {
	g.wmu.Lock()
	defer g.wmu.Unlock()
	return g.write.Encode(g.name, data)
}

func (g *gorillaTokenizer) Decode(data any, token string) (err error) {
	g.rmu.Lock()
	err = securecookie.DecodeMulti(g.name, token, data, g.current, g.previous)
	g.rmu.Unlock()
	if err == nil {
		return nil
	}

	decodingError, ok := err.(interface {
		IsDecode() bool
	})
	// var decodingError securecookie.Error
	// var decodingError securecookie.MultiError
	// if !errors.As(decodingError, err) || !decodingError.IsDecode() {
	if ok && decodingError.IsDecode() {
		// panic(decodingError.IsDecode())
		return nil
	}
	// log.Printf("decoding %x %x", &c.current, &c.previous)
	return err
}

func (g *gorillaTokenizer) Rotate(t time.Time) error {
	fresh := securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)
	// TODO: update KV store with fresh codec.

	g.rmu.Lock()
	g.previous = g.current
	g.current = fresh
	g.rmu.Unlock()

	g.wmu.Lock()
	g.write = fresh
	g.wmu.Unlock()
	return nil
}

type Key struct {
	Label   []byte
	Secret  []byte
	Expires int64
}

// Validate return true if given bytes begin with hex-encoded
// signature that matches HMAC sum of [Key.Label] followed
// by the rest of the data.
func (k *Key) Validate(b []byte) bool {
	// if len(b) <= tokenTagSize {
	// 	return false
	// }
	// if !bytes.Equal(k.Label, b[:12]) {
	// 	// does not match key label
	// 	return false
	// }
	// sha256 block size is 64, but signature length is 32 bytes
	mac := hmac.New(sha256.New, k.Secret)
	_, err := mac.Write(b[tokenSignatureSize:])
	if err != nil {
		return false
	}
	signature := make([]byte, 64)
	_ = hex.Encode(signature, mac.Sum(nil)) // H(key ∥ H(key ∥ message))
	return hmac.Equal(b[:tokenSignatureSize], signature)
}

func (k *Key) Sign(b []byte) []byte {
	mac := hmac.New(sha256.New, k.Secret)
	_, err := mac.Write(k.Label)
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
	return slices.Concat(signature, k.Label, b)
}

func NewKey(d time.Duration) (k *Key, err error) {
	k = &Key{
		Label:   FastRandom(tokenLabelSize),
		Secret:  make([]byte, tokenSignatureSize),
		Expires: time.Now().Add(d).Unix(),
	}
	_, err = io.ReadAtLeast(rand.Reader, k.Secret[:], tokenSignatureSize)
	if err != nil {
		return nil, err
	}
	return k, nil
}

type hmacTokenizer struct {
	wmu   *sync.Mutex
	write *Key

	rmu      *sync.Mutex
	current  *Key
	previous *Key
}

func NewTokenizer() Tokenizer {
	key, err := NewKey(time.Second * 5) // TODO: fix.
	if err != nil {
		panic(err)
	}
	return &hmacTokenizer{
		wmu:   &sync.Mutex{},
		write: key,

		rmu:      &sync.Mutex{},
		current:  key,
		previous: key,
	}
}

func (h *hmacTokenizer) Encode(data any) (string, error) {
	b := &bytes.Buffer{}
	if err := gob.NewEncoder(b).Encode(data); err != nil {
		return "", err
	}
	encoded := make([]byte, base64.URLEncoding.EncodedLen(b.Len()))
	base64.URLEncoding.Encode(encoded, b.Bytes())

	h.wmu.Lock()
	defer h.wmu.Unlock()
	return string(h.write.Sign(encoded)), nil
}

func (h *hmacTokenizer) Decode(data any, token string) (err error) {
	b := []byte(token)
	h.rmu.Lock()
	if !h.current.Validate(b) && !h.previous.Validate(b) {
		// TODO: // OPTIMIZE validation
		return nil
	}
	h.rmu.Unlock()
	b = b[tokenTagSize:] // chomp off tag
	dbuf := make([]byte, base64.URLEncoding.DecodedLen(len(b)))
	if _, err = base64.URLEncoding.Decode(dbuf, b); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewReader(dbuf)).Decode(data)
}

func (h *hmacTokenizer) Rotate(t time.Time) error {
	fresh, err := NewKey(time.Second * 5) // TODO: update
	if err != nil {
		return err
	}
	// TODO: update KV store with fresh codec.

	h.rmu.Lock()
	h.previous = h.current
	h.current = fresh
	h.rmu.Unlock()

	h.wmu.Lock()
	h.write = fresh
	h.wmu.Unlock()
	return nil
}
