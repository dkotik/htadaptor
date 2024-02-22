package secrets

import (
	"errors"
	"log/slog"
	"time"
)

type options struct {
	idSize      int
	entropySize int
	expiry      time.Duration
	window      time.Duration
	logger      *slog.Logger
}

type Option func(*options) error

func WithIDSize(length int) Option {
	return func(o *options) error {
		if length < 4 {
			return errors.New("cannot use ID size smaller than 4 bytes")
		}
		if o.idSize > 4 {
			return errors.New("ID size is already set")
		}
		o.idSize = length
		return nil
	}
}

func WithDefaultIDSizeOfSix() Option {
	return func(o *options) error {
		if o.idSize > 4 {
			return nil
		}
		o.idSize = 6
		return nil
	}
}

func WithEntropySize(length int) Option {
	return func(o *options) error {
		if length < 16 {
			return errors.New("cannot use entropy size smaller than 16 bytes")
		}
		if o.entropySize >= 16 {
			return errors.New("entropy size is already set")
		}
		o.entropySize = length
		return nil
	}
}

func WithDefaultEntropySizeOf32() Option {
	return func(o *options) error {
		if o.entropySize >= 16 {
			return nil
		}
		o.entropySize = 32
		return nil
	}
}

func WithExpiry(d time.Duration) Option {
	return func(o *options) error {
		if d < time.Second {
			return errors.New("cannot use expiration of less than one second")
		}
		if o.expiry >= time.Second {
			return errors.New("expiry is already set")
		}
		o.expiry = d
		return nil
	}
}

func WithDefaultExpiryOfOneWeek() Option {
	return func(o *options) error {
		if o.expiry >= time.Second {
			return nil
		}
		o.expiry = time.Hour * 24 * 7
		return nil
	}
}

func WithRotationWindow(d time.Duration) Option {
	return func(o *options) error {
		if d < time.Second {
			return errors.New("cannot use rotation window of less than one second")
		}
		if o.window >= time.Second {
			return errors.New("rotation window is already set")
		}
		o.window = d
		return nil
	}
}

func WithDefaultRotationWindow() Option {
	return func(o *options) error {
		if o.window >= time.Second {
			return nil
		}
		o.window = o.expiry / 6
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(o *options) error {
		if logger == nil {
			return errors.New("cannot use a <nil> logger")
		}
		if o.logger != nil {
			return errors.New("logger is already set")
		}
		o.logger = logger
		return nil
	}
}

func WithDefaultLogger() Option {
	return func(o *options) error {
		if o.logger != nil {
			return nil
		}
		o.logger = slog.Default()
		return nil
	}
}
