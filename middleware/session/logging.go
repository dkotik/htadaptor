package session

import (
	"context"
	"log/slog"
)

type SlogHandler struct {
	handler slog.Handler
}

func NewSlogHandler(h slog.Handler) *SlogHandler {
	if lh, ok := h.(*SlogHandler); ok {
		h = lh.handler
	}
	return &SlogHandler{h}
}

func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *SlogHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx == nil {
		return h.handler.Handle(ctx, r)
	}
	_ = Read(ctx, func(s Session) error {
		role := s.Role()
		if role == "" {
			role = "guest"
		}
		r.AddAttrs(
			slog.Attr{
				Key:   "session_id",
				Value: slog.StringValue(s.ID()),
			},
			slog.Attr{
				Key:   "trace_id",
				Value: slog.StringValue(s.TraceID()),
			},
			slog.Attr{
				Key:   "user_id",
				Value: slog.StringValue(s.UserID()),
			},
			slog.Attr{
				Key:   "ip_address",
				Value: slog.StringValue(s.Address()),
			},
			slog.Attr{
				Key:   "role",
				Value: slog.StringValue(role),
			},
		)
		return nil
	})
	return h.handler.Handle(ctx, r)
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewSlogHandler(h.handler.WithAttrs(attrs))
}

func (h *SlogHandler) WithGroup(name string) slog.Handler {
	return NewSlogHandler(h.handler.WithGroup(name))
}
