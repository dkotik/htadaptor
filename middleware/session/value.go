package session

import (
	"context"
	"time"
)

func Value(ctx context.Context, key string) any {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Get(key)
}

func SetValue(ctx context.Context, key string, value any) error {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set(key, value)
	return c.encoder.Encode(c.w, c.r, c.values, c.Expires())
}

func ID(ctx context.Context) string {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ID()
}

func TraceID(ctx context.Context) string {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.TraceID()
}

func Role(ctx context.Context) string {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Role()
}

func Address(ctx context.Context) string {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Address()
}

func Expires(ctx context.Context) time.Time {
	c, ok := ctx.Value(contextKey).(*sessionContext)
	if !ok {
		return time.Time{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Expires()
}
