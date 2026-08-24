/*
Package main demonstrates an implementation of a standard
feedback form with validation and localization.
*/
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/middleware/acceptlanguage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type mockSender struct {
	*Letter
}

func (m *mockSender) Send(ctx context.Context, l *Letter) error {
	*m.Letter = *l
	return nil
}

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	bundle := i18n.NewBundle(language.English)
	// mux := http.NewServeMux()

	// mux.Handle("GET /{$}", getForm)
	// mux.Handle("POST /{$}", postForm)

	fmt.Printf(
		`Listening at http://%[1]s/
`,
		l.Addr(),
	)

	letter := new(Letter)
	http.Serve(l, acceptlanguage.New(bundle)(
		htadaptor.Must(NewContactForm(&mockSender{Letter: letter})),
	))
}
