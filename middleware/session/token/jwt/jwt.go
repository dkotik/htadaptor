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
	// panic(fmt.Sprintf("%+v", claims))
	if exp, ok := claims["expires"]; ok { // do not type-cast
		claims["exp"] = exp
	} else if _, ok := claims["exp"]; !ok { // do not type-cast
		return "", errors.New("token must include an `expires` or `exp` field")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	h.wmu.Lock()
	defer h.wmu.Unlock()
	token.Header["kid"] = string(h.write.ID)
	return token.SignedString(h.write.Entropy)
}

func (h *Tokenizer) Decode(data any, tokenString string) (err error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
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
		},
		// jwt.WithJSONNumber(), // bad
		jwt.WithExpirationRequired(),
		jwt.WithPaddingAllowed(),
	)
	if err != nil {
		// TODO: ignoring decoding error
		// panic(err)
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// panic(fmt.Sprintf("%+v", token.Claims))
		// value := reflect.Indirect(reflect.ValueOf(data))
		// value.Set(reflect.MakeMap(value.Type()))
		//
		// target := *(data.(*map[string]any))
		// target["id"] = claims["exp"]
		// (*(data.(*map[string]any)))["id"] = "124"
		// if value.Type().Kind() != reflect.Pointer {
		//   dec.err = errors.New("gob: attempt to decode into a non-pointer")
		//   return dec.err
		// }

		// gob.Register(json.Number("0"))
		// b := &bytes.Buffer{}
		// if err = gob.NewEncoder(b).Encode(claims); err != nil {
		// 	return err
		// }
		// // panic(string(b.Bytes()))
		// if err = gob.NewDecoder(b).Decode(data); err != nil {
		// 	return err
		// }

		*(data.(*map[string]any)) = claims
		// data = claims
		// fmt.Printf("\n\n%+v\n\n", data)
		// fmt.Printf("\n\n%+v || %T\n\n", claims, claims["id"])
		// result := make(map[string]any)
		// for k, v := range claims {
		// 	result[k] = v
		// }
		// *(data.(*map[string]any)) = map[string]any{
		// 	"id": "test",
		// }
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
