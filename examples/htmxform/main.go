/*
Package main demonstrates an implementation of a standard
feedback form with validation and localization.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/middleware/acceptlanguage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type mockSender struct {
	*Letter
	Attempts int
}

func (m *mockSender) Send(ctx context.Context, l *Letter) error {
	if m.Attempts > 0 {
		m.Attempts--
		return errors.New("mock attempt failed, try again")
	}
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

	letter := new(Letter)
	fmt.Printf("http://%[1]s/\n", l.Addr())
	http.Serve(l, acceptlanguage.New(bundle)(
		// slow responses down to show loading indicators
		slowHandler{
			Handler: htadaptor.Must(NewContactForm(
				&mockSender{
					Letter:   letter,
					Attempts: 2,
				},
			)),
		},
	))
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
