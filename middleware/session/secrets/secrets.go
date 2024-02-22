/*
Package secrets provides a synchronizable set of keys
backed by a key-value store.
*/
package secrets

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"
)

type Callback func(present, past *Secret) error

type Secret struct {
	ID      []byte
	Entropy []byte
	Expires int64
}

type Rotation struct {
	Snapshot
	idSize      int
	entropySize int
	expiry      time.Duration
	window      time.Duration
	logger      *slog.Logger
	callback    Callback
}

func NewRotation(callback Callback, withOptions ...Option) (err error) {
	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultIDSizeOfSix(),
		WithDefaultEntropySizeOf32(),
		WithDefaultExpiryOfOneWeek(),
		WithDefaultRotationWindow(),
		WithDefaultLogger(),
		func(o *options) error {
			if callback == nil {
				return errors.New("cannot use a <nil> callback function")
			}
			if o.window < o.expiry/10 {
				return errors.New("rotation window is too short for service stability")
			}
			return nil
		},
	) {
		if err = option(o); err != nil {
			return fmt.Errorf("cannot create a secrets rotation: %w", err)
		}
	}
	r := &Rotation{
		Snapshot:    Snapshot{},
		idSize:      o.idSize,
		entropySize: o.entropySize,
		expiry:      o.expiry,
		window:      o.window,
		logger:      o.logger,
		callback:    callback,
	}
	now := time.Now()
	r.Snapshot.Future, err = r.Next(now.Add(r.expiry).Unix())
	if err != nil {
		return err
	}
	r.Snapshot.Present = r.Snapshot.Future
	if err = r.Rotate(now); err != nil {
		return fmt.Errorf("failed initial key rotation: %w", err)
	}
	go r.Loop(context.Background()) // TODO: add ctx option.
	return nil
}

func (r *Rotation) Rotate(now time.Time) error {
	// if r.Snapshot.Present.Expires > now.Add(r.window).Unix() {
	//   return nil // present key is within the window
	// }
	future, err := r.Next(now.Add(r.expiry * 2).Unix())
	if err != nil {
		return err
	}
	// try time shift
	if err = r.callback(r.Snapshot.Future, r.Snapshot.Present); err != nil {
		return err
	}
	r.Snapshot.Past = r.Snapshot.Present
	r.Snapshot.Present = r.Snapshot.Future
	r.Snapshot.Future = future
	return nil
}

func (r *Rotation) Loop(ctx context.Context) {
	at := time.Now()
	step := time.After(
		time.Unix(r.Snapshot.Present.Expires, 0).Add(-r.window).Sub(at),
	)
	var err error

	for {
		select {
		case <-ctx.Done():
			return
		case at = <-step:
			if err = r.Rotate(at); err != nil {
				r.logger.ErrorContext(
					ctx,
					"key rotation failed",
					slog.Any("error", err),
				)
				step = time.After(time.Minute * 5) // retry sooner
			} else {
				step = time.After(
					time.Unix(r.Snapshot.Present.Expires, 0).Add(-r.window).Sub(at),
				)
				r.logger.DebugContext(
					ctx,
					"performed secrets rotation",
					slog.String("present_secret_id", string(r.Snapshot.Present.ID)),
					slog.String("past_secret_id", string(r.Snapshot.Past.ID)),
				)
			}
		}
	}
}

func (r *Rotation) Next(expires int64) (s *Secret, err error) {
	// TODO: check for ID or entropy collision!
	s = &Secret{
		ID:      NewID(r.idSize),
		Entropy: make([]byte, r.entropySize),
		Expires: expires,
	}
	if _, err = io.ReadAtLeast(
		rand.Reader,
		s.Entropy,
		r.entropySize,
	); err != nil {
		return nil, err
	}
	return s, nil
}
