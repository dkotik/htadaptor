package session

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

func Read(ctx context.Context, view func(Session) error) (err error) {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return errors.New("no session context")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.values == nil {
		_ = c.decoder.Decode(&c.values, c.w, c.r)
		if c.values == nil {
			c.values = c.factory()
			if err = view(c); err != nil {
				return err
			}
			return c.encoder.Encode(c.w, c.r, c.values)
		}
	}
	return view(c)
}

func Write(ctx context.Context, update func(Session) error) (err error) {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return errors.New("no session context")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.values == nil {
		err = c.decoder.Decode(&c.values, c.w, c.r)
		if err != nil && !errors.Is(http.ErrNoCookie, err) {
			decodingError, ok := err.(interface {
				IsDecode() bool
			})
			// var decodingError securecookie.Error
			// var decodingError securecookie.MultiError
			// if !errors.As(decodingError, err) || !decodingError.IsDecode() {
			if !ok || !decodingError.IsDecode() {
				// panic(decodingError.IsDecode())
				return err
			}
		}
		if c.values == nil {
			c.values = c.factory()
		}
	}
	if err = update(c); err != nil {
		return err
	}
	return c.encoder.Encode(c.w, c.r, c.values)
}

type sessionContext struct {
	context.Context
	encoder Encoder
	decoder Decoder
	factory Factory

	mu     *sync.Mutex
	w      http.ResponseWriter
	r      *http.Request
	values map[string]any

	id      string
	traceID string
}

func (c *sessionContext) ID() string {
	if c.id == "" {
		c.id, _ = c.values["id"].(string)
	}
	return c.id
}

func (c *sessionContext) TraceID() string {
	if c.traceID == "" {
		c.traceID = FastRandom(8)
	}
	return c.traceID
}

func (c *sessionContext) Get(key string) any {
	value, ok := c.values[key]
	if !ok {
		return nil
	}
	return value
}

func (c *sessionContext) Set(key string, value any) {
	c.values[key] = value
}

// func (c *sessionContext) Close() error {
// 	defer c.mu.Unlock()
// 	if !c.modified {
// 		return nil
// 	}
// 	// panic(c.values)
// 	return
// }

func (c *sessionContext) Reset() {
	c.values = c.factory()
}

func (c *sessionContext) Value(key any) any {
	switch key.(type) {
	case contextKeyType:
		return c
	}
	return c.Context.Value(key)
}
