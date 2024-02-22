/*
Package gorilla couples [securecookie.SecureCookie] token encoder
with [secrets.Rotation].
*/
package gorilla

import (
	"sync"

	"github.com/dkotik/htadaptor/middleware/session/secrets"
	"github.com/gorilla/securecookie"
)

type Tokenizer struct {
	name string

	wmu   *sync.Mutex
	write *securecookie.SecureCookie

	rmu      *sync.Mutex
	current  *securecookie.SecureCookie
	previous *securecookie.SecureCookie
}

func New(name string, withOptions ...secrets.Option) *Tokenizer {
	t := &Tokenizer{
		name: name,
		wmu:  &sync.Mutex{},
		rmu:  &sync.Mutex{},
	}
	// size is important!
	withOptions = append(withOptions, secrets.WithEntropySize(32+64))
	if err := secrets.NewRotation(t.Rotate, withOptions...); err != nil {
		panic(err) // TODO: beautify.
	}
	return t
}

func (g *Tokenizer) Encode(data any) (string, error) {
	g.wmu.Lock()
	defer g.wmu.Unlock()
	return g.write.Encode(g.name, data)
}

func (g *Tokenizer) Decode(data any, token string) (err error) {
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

func (g *Tokenizer) Rotate(present, past *secrets.Secret) error {
	// securecookie.GenerateRandomKey(64),
	// securecookie.GenerateRandomKey(32),
	presentSC := securecookie.New(present.Entropy[:64], present.Entropy[64:])
	pastSC := securecookie.New(past.Entropy[:64], past.Entropy[64:])

	g.rmu.Lock()
	g.previous = pastSC
	g.current = presentSC
	g.rmu.Unlock()

	g.wmu.Lock()
	g.write = presentSC
	g.wmu.Unlock()
	return nil
}
