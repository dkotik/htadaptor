package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type ContextFactory func(net.Listener) context.Context

type options struct {
	TLSCertificateFile string
	TLSKeyFile         string
	ReadTimeout        time.Duration
	ReadHeaderTimeout  time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	MaxHeaderBytes     int
	Logger             *slog.Logger
	Listener           net.Listener
	ContextFactory     ContextFactory
	Handler            http.Handler
}

type Option func(*options) error

func WithDebugOptions() Option {
	return func(o *options) (err error) {
		if o.Logger == nil {
			if err = WithLogger(NewDebugLogger())(o); err != nil {
				return err
			}
		}
		if o.Listener == nil {
			if err = WithAddress("localhost", 8080)(o); err != nil {
				return err
			}
		}
		return nil
	}
}

func WithDefaultOptions() Option {
	return func(o *options) (err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("cannot apply defaults: %w", err)
			}
		}()

		if o.ReadTimeout == 0 {
			if err = WithReadTimeout(time.Second * 2)(o); err != nil {
				return err
			}
		}
		if o.ReadHeaderTimeout == 0 {
			if err = WithReadHeaderTimeout(time.Second * 2)(o); err != nil {
				return err
			}
		}
		if o.WriteTimeout == 0 {
			if err = WithWriteTimeout(time.Second * 2)(o); err != nil {
				return err
			}
		}
		if o.IdleTimeout == 0 {
			if err = WithIdleTimeout(time.Second * 30)(o); err != nil {
				return err
			}
		}
		if o.MaxHeaderBytes == 0 {
			if err = WithMaxHeaderBytes(1 << 6)(o); err != nil {
				return err
			}
		}
		if o.Listener == nil {
			if o.TLSCertificateFile != "" {
				if err = WithAddress("", 443)(o); err != nil {
					return err
				}
			} else {
				if err = WithAddress("", 8080)(o); err != nil {
					return err
				}
			}
		}
		if o.Logger == nil {
			if err = WithLogger(slog.Default().With(
				slog.String("commit", vcsCommit()),
			))(o); err != nil {
				return err
			}
		}
		if o.Handler == nil {
			if err = WithHandler(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Empty Handler", http.StatusNotFound)
				},
			))(o); err != nil {
				return err
			}
		}
		return nil
	}
}

func WithReadTimeout(t time.Duration) Option {
	return func(o *options) error {
		if o.ReadTimeout != 0 {
			return errors.New("read timeout is already set")
		}
		if t < time.Millisecond*100 {
			return errors.New("cannot set read timeout lower than 100ms")
		}
		if t > time.Minute {
			return errors.New("cannot set read timeout above one minute")
		}
		o.ReadTimeout = t
		return nil
	}
}

func WithReadHeaderTimeout(t time.Duration) Option {
	return func(o *options) error {
		if o.ReadHeaderTimeout != 0 {
			return errors.New("read header timeout is already set")
		}
		if t < time.Millisecond*20 {
			return errors.New("cannot set read header timeout lower than 20ms")
		}
		if t > time.Second*10 {
			return errors.New("cannot set read header timeout above ten seconds")
		}
		o.ReadHeaderTimeout = t
		return nil
	}
}

func WithWriteTimeout(t time.Duration) Option {
	return func(o *options) error {
		if o.WriteTimeout != 0 {
			return errors.New("write timeout is already set")
		}
		if t < time.Millisecond*100 {
			return errors.New("cannot set write timeout lower than 100ms")
		}
		if t > time.Minute {
			return errors.New("cannot set write timeout above one minute")
		}
		o.WriteTimeout = t
		return nil
	}
}

func WithIdleTimeout(t time.Duration) Option {
	return func(o *options) error {
		if o.IdleTimeout != 0 {
			return errors.New("idle timeout is already set")
		}
		if t < time.Millisecond*100 {
			return errors.New("cannot set idle timeout lower than 100ms")
		}
		if t > time.Minute {
			return errors.New("cannot set idle timeout above one minute")
		}
		o.IdleTimeout = t
		return nil
	}
}

func WithMaxHeaderBytes(m int) Option {
	return func(o *options) error {
		if o.MaxHeaderBytes != 0 {
			return errors.New("maximum header bytes limit is already set")
		}
		if m < 1 {
			return errors.New("cannot read less than 1 header bytes")
		}
		if m > 1<<24 {
			return errors.New("max header bytes is too large")
		}
		o.MaxHeaderBytes = m
		return nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(o *options) error {
		if logger == nil {
			return errors.New("cannot use a <nil> structured logger")
		}
		o.Logger = logger
		return nil
	}
}

func WithListener(l net.Listener) Option {
	return func(o *options) error {
		if o.Listener != nil {
			return errors.New("address is already set")
		}
		if l == nil {
			return errors.New("cannot use a <nil> network listener")
		}
		o.Listener = l
		return nil
	}
}

func WithAddress(host string, port uint32) Option {
	return func(o *options) (err error) {
		if port < 1 {
			return errors.New("cannot use port lower than 1")
		}
		address := fmt.Sprintf("%s:%d", host, port)
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return fmt.Errorf("cannot bind listener to %q address: %w", address, err)
		}
		return WithListener(listener)(o)
	}
}

func WithContextFactory(f ContextFactory) Option {
	return func(o *options) error {
		if o.ContextFactory != nil {
			return errors.New("context factory is already set")
		}
		if f == nil {
			return errors.New("cannot use a <nil> context factory")
		}
		o.ContextFactory = f
		return nil
	}
}

func WithTraceIDGenerator(generator func() string) Option {
	return func(o *options) error {
		if generator == nil {
			return errors.New("cannot use a <nil> trace id generator")
		}
		return WithContextFactory(func(_ net.Listener) context.Context {
			return ContextWithTraceIDGenerator(context.Background(), generator)
		})(o)
	}
}

func WithDefaultTraceIDGenerator() Option {
	return func(o *options) error {
		if o.ContextFactory == nil {
			return WithTraceIDGenerator(func() string {
				return FastRandom(8)
			})(o)
		}
		return nil
	}
}

func WithTLS(certificateFile, keyFile string) Option {
	return func(o *options) error {
		if o.TLSCertificateFile != "" {
			return errors.New("TLS option is already set")
		}
		o.TLSCertificateFile = certificateFile
		o.TLSKeyFile = keyFile
		return nil
	}
}

func WithHandler(h http.Handler) Option {
	return func(o *options) error {
		if o.Handler != nil {
			return errors.New("HTTP handler is already set")
		}
		if h == nil {
			return errors.New("cannot use a <nil> HTTP handler")
		}
		o.Handler = h
		return nil
	}
}
