package service

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type contextKey struct{}

type Traceable interface {
	GetTraceID() string
}

type immediateTracing struct {
	id string
}

func (i *immediateTracing) GetTraceID() string {
	return i.id
}

type lazyTracing struct {
	generator func() string

	mu sync.Mutex
	id string
}

func (l *lazyTracing) GetTraceID() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.id == "" {
		l.id = l.generator()
	}
	return l.id
}

func ContextWithTraceIDGenerator(parent context.Context, generator func() string) context.Context {
	if generator == nil {
		generator = func() string {
			return ""
		}
	}
	return ContextWithTracing(parent, &lazyTracing{
		generator: generator,
		mu:        sync.Mutex{},
	})
}

func ContextWithTraceID(parent context.Context, ID string) context.Context {
	return ContextWithTracing(parent, &immediateTracing{id: ID})
}

func ContextWithTracing(parent context.Context, t Traceable) context.Context {
	return context.WithValue(parent, contextKey{}, t)
}

func TraceIDFromContext(ctx context.Context) string {
	t, _ := ctx.Value(contextKey{}).(Traceable)
	if t == nil {
		return ""
	}
	return t.GetTraceID()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// FastRandom is inspired by Ketan Parmar's work:
//
// - https://github.com/kpbird/golang_random_string/blob/master/main.go
// - https://kpbird.medium.com/golang-generate-fixed-size-random-string-dd6dbd5e63c0
func FastRandom(n int) string {
	b := make([]byte, n)
	l := len(letterBytes)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < l {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
