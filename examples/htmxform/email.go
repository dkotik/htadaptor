package main

import (
	"context"
	"log/slog"

	"github.com/dkotik/htadaptor/examples/htmxform/feedback"
)

// in real application replace with adaptor to mail provider
var mailer = feedback.Sender(
	func(ctx context.Context, r *feedback.Request) error {
		slog.Default().InfoContext(
			ctx,
			"received a feedback request",
			slog.Any("request", r),
		)
		return nil
	},
)
