package session

import (
	"context"
	"net/http"
	"sync"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

func Lock(ctx context.Context) (s Session, unlock func()) {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return nil, nil
	}
	c.mu.Lock()
	if c.values == nil {
		// c.id = base32RawStdEncoding.EncodeToString(
		// 	securecookie.GenerateRandomKey(32))
		// var data any
		_ = c.decoder.Decode(&c.values, c.w, c.r)
		// TODO deal with error?
		// if err != nil {
		// 	fmt.Println("decoding failure:", err.Error())
		// }
		if c.values == nil {
			c.values = map[string]any{
				"id": FastRandom(32),
			}
			return c, func() {
				_ = c.encoder.Encode(c.w, c.r, c.values)
				// TODO deal with error?
				c.mu.Unlock()
			}
		}
	}
	return c, c.mu.Unlock
}

type sessionContext struct {
	context.Context
	encoder Encoder
	decoder Decoder

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
	if key == "" || key == "id" {
		// TODO: remove panic.
		panic("invalid session data key: no empty of \"id\" keys")
	}
	c.values[key] = value
}

func (c *sessionContext) Commit() error {
	return c.encoder.Encode(c.w, c.r, c.values)
}

func (c *sessionContext) Rotate() error {
	return nil
}

func (c *sessionContext) Value(key any) any {
	switch key.(type) {
	case contextKeyType:
		return c
	}
	return c.Context.Value(key)
}
