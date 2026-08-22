/*
Package main demonstrates application of generic domain adaptors to satisfy [http.Handler] interface.
*/
package main

import (
	"net/http"
	"os"

	"log/slog"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/extract"
)

const (
	pathOrder     = "/api/v1/order/{number}"
	pathPrice     = "/api/v1/price"
	pathInventory = "/api/v1/inventory"
)

func newService(store *OnlineStore) *http.ServeMux {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	adaptor := htadaptor.New()
	mux := http.NewServeMux()
	mux.Handle(pathOrder, htadaptor.Must(
		adaptor.AdaptFunc(
			store.Order,
			htadaptor.WithPathValues("number"),
		),
	))

	mux.Handle(pathPrice, htadaptor.Must(
		adaptor.AdaptStringFunc(
			store.GetPrice,
			extract.StringValueExtractorFunc(
				func(r *http.Request) (string, error) {
					return r.URL.Query().Get("item"), nil
				},
			),
		),
	))

	mux.Handle(pathInventory, htadaptor.Must(
		adaptor.AdaptNullaryFunc(store.GetInventory),
	))
	return mux
}

func main() {
	http.ListenAndServe(":0", newService(&OnlineStore{}))
}
