/*
Package main demonstrates application of generic domain adaptors to satisfy [http.Handler] interface.
*/
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"log/slog"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/extractor"
)

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	domainLogic := &OnlineStore{}
	mux := http.NewServeMux()
	mux.Handle("/api/v1/order/{number}", htadaptor.Must(
		htadaptor.NewUnaryFuncAdaptor(
			domainLogic.Order,
			htadaptor.WithPathValues("number"),
		),
	))

	mux.Handle("/api/v1/price", htadaptor.Must(
		htadaptor.NewUnaryStringFuncAdaptor(
			domainLogic.GetPrice,
			extractor.StringValueExtractorFunc(
				func(r *http.Request) (string, error) {
					log.Println("got query:", r.URL.RawQuery)
					return r.URL.Query().Get("item"), nil
				},
			),
		),
	))

	mux.Handle("/api/v1/inventory", htadaptor.Must(
		htadaptor.NewNullaryFuncAdaptor(domainLogic.GetInventory),
	))

	mux.Handle("/api/v1/record", htadaptor.Must(
		htadaptor.NewVoidFuncAdaptor(domainLogic.Record),
	))

	fmt.Printf(
		`Listening at http://%[1]s/

    Test Order (Unary):
      curl -v -d '{"item":"box","quantity":1}' -H 'Content-Type: application/json' http://%[1]s/api/v1/order/1
    Test Price (Unary String):
      curl -v -G -d 'item=shirt' http://%[1]s/api/v1/price
    Test Inventory (Nullary):
      curl -v http://%[1]s/api/v1/inventory
    Test Record (Unary Void):
      curl -v -d '{"item":"box","quantity":1}' -H 'Content-Type: application/json' http://%[1]s/api/v1/record

`,
		l.Addr(),
	)

	http.Serve(l, mux)
}
