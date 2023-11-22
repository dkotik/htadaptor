/*
Package main demonstrates routing directly to domain function calls.
*/
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"log/slog"

	"github.com/dkotik/htadaptor"
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
	mux.Handle("/api/v1/order", htadaptor.Must(
		htadaptor.NewUnaryFuncAdaptor(domainLogic.Order),
	))

	mux.Handle("/api/v1/inventory", htadaptor.Must(
		htadaptor.NewNullaryFuncAdaptor(domainLogic.GetInventory),
	))

	mux.Handle("/api/v1/record", htadaptor.Must(
		htadaptor.NewVoidFuncAdaptor(domainLogic.Record),
	))

	// TODO: add string routes.
	// handler, err := oakmux.New(
	// 	oakmux.WithRouteStringFunc( // UnaryString
	// 		"price", "price",
	// 		domainLogic.GetPrice,
	// 		func(r *http.Request) (string, error) {
	// 			// string decoder
	// 			log.Println("got query:", r.URL.RawQuery)
	// 			return r.URL.Query().Get("item"), nil
	// 		},
	// 	),
	// )

	fmt.Printf(
		`Listening at http://%[1]s/

    Test Order (Unary):
      curl -v -d '{"item":"box","quantity":1}' -H 'Content-Type: application/json' http://%[1]s/api/v1/order
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
