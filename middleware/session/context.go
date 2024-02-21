package session

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

const (
	expiresField = "expires"
	roleField    = "role"
)

func Read(ctx context.Context, view func(Session) error) (err error) {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return errors.New("no session context")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.values == nil {
		if err = c.decoder.Decode(&c.values, c.w, c.r); err != nil {
			return err
		}
		if c.values == nil || c.IsExpired() {
			c.values = c.factory()
			if err = view(c); err != nil {
				return err
			}
			return c.encoder.Encode(c.w, c.r, c.values, c.Expires())
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
		if err = c.decoder.Decode(&c.values, c.w, c.r); err != nil {
			return err
		}
		if c.values == nil || c.IsExpired() {
			c.values = c.factory()
		}
	}
	if err = update(c); err != nil {
		return err
	}
	return c.encoder.Encode(c.w, c.r, c.values, c.Expires())
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

func (c *sessionContext) Role() (s string) {
	s, _ = c.values[roleField].(string)
	return s
}

func (c *sessionContext) Expires() (t time.Time) {
	t, _ = c.values[expiresField].(time.Time)
	return t
}

func (c *sessionContext) IsExpired() bool {
	expires, ok := c.values[expiresField].(int64)
	return !ok || expires <= time.Now().Unix()
}

func (c *sessionContext) Address() string {
	return c.r.RemoteAddr
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
