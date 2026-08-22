/*
Package main demonstrates the use of template to create an HTMX endpoint.
*/
package main

import (
	"context"
	"html/template"
	"net/http"

	"github.com/dkotik/htadaptor"
)

type testResponse struct {
	Name string
}

var greetingTemplate = template.Must(
	template.New("greeting").Parse(`
    <h1>Hello {{ .Name }}!</h1>
    <p>Enjoy this HTMX component!</p>
  `),
)

func newHandlerHTMX() (http.Handler, error) {
	// see examples/htmxform for an interactive HTMX component
	adaptor := htadaptor.New()
	return adaptor.AdaptNullaryFunc(
		func(ctx context.Context) (*testResponse, error) {
			return &testResponse{
				Name: "Guest",
			}, nil
		},
		htadaptor.WithTemplate(greetingTemplate),
	)
}

const handlerPath = "/test/htmx"

func main() {
	mux := http.NewServeMux()
	mux.Handle(
		handlerPath,
		htadaptor.Must(newHandlerHTMX()),
	)

	http.ListenAndServe(":0", mux)
}
