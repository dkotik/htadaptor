package session

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/dkotik/htadaptor/middleware/session/secrets"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

const (
	expiresField = "expires"
	roleField    = "role"
	userField    = "user_id"
)

func Read(ctx context.Context, view func(Session) error) (err error) {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return ErrNoSessionInContext
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.values == nil {
		if err = c.readCookieToken(); err != nil {
			return err
		}
		if c.values == nil || c.IsExpired() {
			c.Reset()
			if err = view(c); err != nil {
				return err
			}
			return c.writeCookieToken()
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
		if err = c.readCookieToken(); err != nil {
			return err
		}
		if c.values == nil || c.IsExpired() {
			c.Reset()
		}
	}
	if err = update(c); err != nil {
		return err
	}
	return c.writeCookieToken()
}

type sessionContext struct {
	context.Context
	cookies   CookieCodec
	tokenizer Tokenizer
	factory   Factory

	mu     *sync.Mutex
	w      http.ResponseWriter
	r      *http.Request
	values map[string]any

	id      string
	traceID string
	isNew   bool
}

func (c *sessionContext) readCookieToken() error {
	cookie := c.cookies.ReadCookie(c.r)
	if cookie == "" {
		return nil
	}
	return c.tokenizer.Decode(&c.values, cookie)
}

func (c *sessionContext) writeCookieToken() error {
	token, err := c.tokenizer.Encode(c.values)
	if err != nil {
		return err
	}
	return c.cookies.WriteCookie(c.w, token, c.Expires())
}

func (c *sessionContext) ID() string {
	if c.id == "" {
		c.id, _ = c.values["id"].(string)
	}
	return c.id
}

func (c *sessionContext) TraceID() string {
	if c.traceID == "" {
		c.traceID = string(secrets.NewID(8))
	}
	return c.traceID
}

func (c *sessionContext) Role() (s string) {
	s, _ = c.values[roleField].(string)
	return s
}

func (c *sessionContext) SetRole(name string) {
	c.values[roleField] = name
}

func (c *sessionContext) UserID() (s string) {
	s, _ = c.values[userField].(string)
	return s
}

func (c *sessionContext) SetUserID(id string) {
	c.values[userField] = id
}

func (c *sessionContext) Expires() time.Time {
	t, _ := c.values[expiresField].(int64)
	return time.Unix(t, 0)
}

func (c *sessionContext) IsExpired() bool {
	expires, ok := c.values[expiresField].(int64)
	return !ok || expires <= time.Now().Unix()
}

func (c *sessionContext) IsNew() bool {
	return c.isNew
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
	c.isNew = true
}

func (c *sessionContext) Value(key any) any {
	switch key.(type) {
	case contextKeyType:
		return c
	}
	return c.Context.Value(key)
}
