package service

import (
	"context"
	"log/slog"
	"math"
	"os"
	"runtime/debug"
	"time"

	"github.com/lmittmann/tint"
)

func vcsCommit() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, kv := range info.Settings {
			switch kv.Key {
			case "vcs.revision":
				return kv.Value
			}
		}
	}
	return "<unknown>"
}

type tracingHandler struct {
	slog.Handler
}

func (t *tracingHandler) Handle(ctx context.Context, r slog.Record) error {
	if ID := TraceIDFromContext(ctx); ID != "" {
		r.AddAttrs(slog.String("traceID", ID))
	}
	return t.Handler.Handle(ctx, r)
}

func NewTracingHandler(h slog.Handler) slog.Handler {
	return &tracingHandler{
		Handler: h,
	}
}

func NewDebugLogger() *slog.Logger {
	return slog.New(NewTracingHandler(
		tint.NewHandler(os.Stderr, &tint.Options{
			// Level:      slog.LevelDebug,
			Level:      slog.Level(-math.MaxInt), // log everything
			TimeFormat: time.Kitchen,
		}))).With(
		slog.String("commit", vcsCommit()),
	)
}

type slogAdaptor struct {
	logger *slog.Logger
	level  slog.Level
}

func (s *slogAdaptor) Write(b []byte) (n int, err error) {
	s.logger.Log(context.Background(), s.level, string(b))
	return len(b), nil
}
