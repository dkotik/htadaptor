/*
Package main demonstrates an implementation of a standard
feedback form with validation and localization.
*/
package main

import (
	"context"
	_ "embed" // for spinner
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/middleware/acceptlanguage"
	"github.com/dkotik/htadaptor/staticfs"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed spinner.svg
var spinner []byte

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
	mux := http.NewServeMux()
	mux.Handle("GET /spinner.svg", staticfs.NewFastFileSystemFileWithContentType(spinner, "image/svg+xml"))

	letter := new(Letter)
	mux.Handle("/{$}", acceptlanguage.New(bundle)(
		slowHandler{ // slow responses down to show loading indicators
			Handler: htadaptor.Must(NewContactForm(&mockSender{Letter: letter})),
		},
	))

	fmt.Printf(`Listening at http://%[1]s/ `, l.Addr())
	http.Serve(l, mux)
}

type slowHandler struct {
	http.Handler
}

func (h slowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		return
	case <-time.After(time.Second * 2):
	}
	h.Handler.ServeHTTP(w, r)
}
