/*
Package main demonstrates the use of template to create an HTMX endpoint.
*/
package main

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/dkotik/htadaptor"
)

type testResponse struct {
	Name string
}

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	greetingTemplate, err := template.New("greeting").Parse(`
    <h1>Hello {{ .Name }}!</h1>
    <p>Enjoy this HTMX component!</p>
  `)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/test/htmx", htadaptor.Must(
		htadaptor.NewNullaryFuncAdaptor(
			func(ctx context.Context) (*testResponse, error) {
				return &testResponse{
					Name: "Guest",
				}, nil
			},
			htadaptor.WithTemplate(greetingTemplate),
		),
	))

	fmt.Printf(
		`Listening at http://%[1]s/

    Test HTMX component:
      curl -v http://%[1]s/test/htmx
`,
		l.Addr(),
	)

	http.Serve(l, mux)
}
