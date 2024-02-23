/*
Package session presents a lazy context that manages session state
with native key rotation support.
*/
package session

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Session interface {
	ID() string
	TraceID() string
	Address() string
	UserID() string
	SetUserID(string)
	Role() string
	SetRole(string)
	Expires() time.Time
	IsExpired() bool
	IsNew() bool
	Get(string) any
	Int(string) int
	Int64(string) int64
	Float32(string) float32
	Float64(string) float64
	Set(string, any)
	Reset()
}

func New(withOptions ...Option) (func(http.Handler) http.Handler, error) {
	options := &options{}
	var err error
	for _, option := range append(
		withOptions,
		WithDefaultName(),
		WithDefaultExpiry(),
		WithDefaultRotationContext(),
		WithDefaultTokenizer(),
		WithDefaultCookieCodec(),
		WithDefaultFactory(),
	) {
		if err = option(options); err != nil {
			return nil, fmt.Errorf("cannot create session middleware: %w", err)
		}
	}
	tokenizer := options.Tokenizer
	// go func(ctx context.Context, tokenizer Tokenizer) {
	// 	t := time.NewTicker(options.Expiry * 9 / 10)
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		case at := <-t.C:
	// 			_ = tokenizer.Rotate(at)
	// 			// log.Printf("---------- rotated %x %x", &fresh, &dec.current)
	// 		}
	// 	}
	// }(options.RotationContext, tokenizer)
	factory := options.Factory
	cookieCodec := options.CookieCodec

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(
					&sessionContext{
						Context:   r.Context(),
						cookies:   cookieCodec,
						tokenizer: tokenizer,
						factory:   factory,

						mu: &sync.Mutex{},
						w:  w,
						r:  r,
					},
				))
			},
		)
	}, nil
}
