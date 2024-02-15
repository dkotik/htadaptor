package idledown

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/dkotik/htadaptor/service"
)

func TestIdleDown(t *testing.T) {
	ctx, idleDown := New(context.Background(), time.Second)
	err := service.Run(
		ctx,
		service.WithAddress("localhost", 8989),
		service.WithDebugOptions(),
		service.WithHandler(
			idleDown(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("ok"))
				},
			)),
		),
	)

	if err != nil {
		t.Fatal(err)
	}
	// t.Fatal("over")
}
