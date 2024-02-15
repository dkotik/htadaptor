package service

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	t.Skip("test needs fixing")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := NewDebugLogger()

	go func() {
		err := Run(
			ctx,
			WithHandler(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					logger.Log(r.Context(), slog.LevelInfo, "trace IDs must match between two entries")
					// panic("boo")
				},
			)),
			WithLogger(logger),
			WithDebugOptions(),
			WithAddress("localhost", 8888),
		)
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Millisecond * 100)
	resp, err := http.Get("http://localhost:8888")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatal("unexpected status code:", resp.StatusCode)
	}
}
