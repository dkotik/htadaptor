package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(ctx context.Context, withOptions ...Option) (err error) {
	// 1 	SIGHUP 	Terminate 	Hang up controlling terminal or process. Sometimes used as a signal to reread configuration file for the program.
	// 2 	SIGINT 	Terminate 	Interrupt from keyboard, Ctrl + C.
	// 3 	SIGQUIT 	Dump 	Quit from keyboard, Ctrl + \.
	// 9 	SIGKILL 	Terminate 	Forced termination.
	// 15 	SIGTERM 	Terminate 	Graceful termination.
	// 17 	SIGCHLD 	Ignore 	Child process exited.
	// 18 	SIGCONT 	Continue 	Resume process execution.
	// 19 	SIGSTOP 	Stop 	Stop process execution, Ctrl + Z.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	o := &options{}
	for _, option := range append(
		withOptions,
		WithDefaultOptions(),
		WithDefaultTraceIDGenerator(),
		func(o *options) error { // validate
			if o.Listener == nil {
				return errors.New("cannot start a server without a network listener")
			}
			return nil
		},
	) {
		if err = option(o); err != nil {
			return fmt.Errorf("cannot create an HTTP service: %w", err)
		}
	}
	ln := o.Listener
	defer func() {
		if err := ln.Close(); err != nil {
			o.Logger.Error("failed closing network listener", slog.Any("error", err))
		}
	}()

	handler := o.Handler
	contextFactory := o.ContextFactory
	logger := o.Logger
	server := &http.Server{
		// Addr:              o.Address,
		ReadTimeout:       o.ReadTimeout,
		ReadHeaderTimeout: o.ReadHeaderTimeout,
		WriteTimeout:      o.WriteTimeout,
		IdleTimeout:       o.IdleTimeout,
		MaxHeaderBytes:    o.MaxHeaderBytes,
		Handler:           handler,
		BaseContext:       contextFactory,
		ErrorLog: log.New(&slogAdaptor{
			level:  slog.LevelDebug,
			logger: logger,
		}, "HTTP: ", log.LstdFlags),
		// TLSConfig *tls.Config // TODO: use Filippo's defaults?
	}

	go func(ctx context.Context, logger *slog.Logger) {
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// see join below
		// if err := context.Cause(ctx); err != nil {
		// 	logger.Error("context cancelled", slog.Any("reason", err))
		// }
		if err := server.Shutdown(timeoutCtx); err != nil {
			logger.Error("error shutting down HTTP service", slog.Any("error", err))
		}
	}(ctx, logger)

	if o.TLSCertificateFile != "" {
		// err = server.ListenAndServeTLS(o.TLSCertificateFile, o.TLSKeyFile)
		err = server.ServeTLS(ln, o.TLSCertificateFile, o.TLSKeyFile)
	} else {
		// if strings.HasSuffix(o.Address, ":443") || strings.HasSuffix(o.Address, ":8443") {
		// 	return errors.New("must not expose a TLS server without its certificate file set")
		// }
		err = server.Serve(ln)
	}

	err = errors.Join(err, context.Cause(ctx))
	if err != nil {
		logger.Error("service shutdown", slog.Any("reason", err)) // handle
	}
	return nil
}
