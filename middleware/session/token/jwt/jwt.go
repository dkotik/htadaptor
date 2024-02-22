/*
Package jwt provides [token.Authority] with secure JSON Web Tokens
issuer defaults.
*/
package jwt

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/dkotik/htadaptor/middleware/session/secrets"
	"github.com/golang-jwt/jwt/v5"
)

type Tokenizer struct {
	wmu   *sync.Mutex
	write *secrets.Secret

	rmu     *sync.Mutex
	present *secrets.Secret
	past    *secrets.Secret
}

func New(withOptions ...secrets.Option) *Tokenizer {
	t := &Tokenizer{
		wmu: &sync.Mutex{},
		rmu: &sync.Mutex{},
	}

	if err := secrets.NewRotation(t.Rotate, withOptions...); err != nil {
		panic(err) // TODO: beautify.
	}
	return t
}

func (h *Tokenizer) Encode(data any) (string, error) {
	claims := jwt.MapClaims(data.(map[string]any))
	if exp, ok := claims["expires"].(int64); ok {
		claims["exp"] = exp
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	h.wmu.Lock()
	defer h.wmu.Unlock()
	token.Header["kid"] = string(h.write.ID)
	return token.SignedString(h.write.Entropy)
}

func (h *Tokenizer) Decode(data any, tokenString string) (err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // important
			return nil, fmt.Errorf("signing method other than HMAC-sha256 not supported: %v", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string) // TODO: test without kid
		if !ok {
			return nil, errors.New("token header does not contain a key id")
		}
		id := []byte(kid)

		h.rmu.Lock()
		defer h.rmu.Unlock()
		if bytes.Equal(h.present.ID, id) {
			return h.present.Entropy, nil
		}
		if bytes.Equal(h.past.ID, id) {
			return h.past.Entropy, nil
		}
		return nil, fmt.Errorf("none of the keys matched key id: %s", id)
	})
	if err != nil {
		// TODO: ignoring decoding error
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		*(data.(*map[string]any)) = claims
	}
	return nil
}

func (h *Tokenizer) Rotate(present, past *secrets.Secret) error {
	h.rmu.Lock()
	h.present = present
	h.past = past
	h.rmu.Unlock()

	h.wmu.Lock()
	h.write = present
	h.wmu.Unlock()
	return nil
}
