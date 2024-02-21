package session

import (
	"context"
	"errors"
	"time"
)

type Factory func() map[string]any

type options struct {
	Name            string
	Expiry          time.Duration
	RotationContext context.Context
	Factory         Factory
}

type Option func(*options) error

func WithName(name string) Option {
	return func(o *options) error {
		if name == "" {
			return errors.New("cannot use an empty session name")
		}
		if o.Name != "" {
			return errors.New("session name is already set")
		}
		o.Name = name
		return nil
	}
}

func WithDefaultName() Option {
	return func(o *options) error {
		if o.Name != "" {
			return nil
		}
		o.Name = "rotatingSession"
		return nil
	}
}

func WithExpiry(d time.Duration) Option {
	return func(o *options) error {
		if d < time.Millisecond {
			return errors.New("session expiry cannot be less than a millisecond")
		}
		if o.Expiry > 0 {
			return errors.New("session expiry is already set")
		}
		o.Expiry = d
		return nil
	}
}

func WithDefaultExpiry() Option {
	return func(o *options) error {
		if o.Expiry > 0 {
			return nil
		}
		o.Expiry = time.Hour * 24 * 14
		return nil
	}
}

func WithRotationContext(ctx context.Context) Option {
	return func(o *options) error {
		if ctx == nil {
			return errors.New("cannot use empty rotation context")
		}
		if o.RotationContext != nil {
			return errors.New("rotation context is already set")
		}
		o.RotationContext = ctx
		return nil
	}
}

func WithDefaultRotationContext() Option {
	return func(o *options) error {
		if o.RotationContext != nil {
			return nil
		}
		o.RotationContext = context.Background()
		return nil
	}
}

func WithFactory(f Factory) Option {
	return func(o *options) error {
		if f == nil {
			return errors.New("cannot use empty session factory")
		}
		if o.Factory != nil {
			return errors.New("session factory is already set")
		}
		o.Factory = f
		return nil
	}
}

func WithDefaultFactory() Option {
	return func(o *options) error {
		if o.Factory != nil {
			return nil
		}
		o.Factory = func() map[string]any {
			return map[string]any{
				"id": FastRandom(32),
			}
		}
		return nil
	}
}
