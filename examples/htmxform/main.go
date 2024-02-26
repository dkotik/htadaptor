/*
Package main demonstrates an implementation of a standard
feedback form with validation and localization.
*/
package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/dkotik/htadaptor/examples/htmxform/feedback"
	"github.com/dkotik/htadaptor/middleware/acceptlanguage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	bundle := i18n.NewBundle(language.English)
	feedback.LoadEnglish(bundle)
	feedback.LoadRussian(bundle)

	formEndpoint := "/api/v1/feedback.htmx"
	mux := http.NewServeMux()
	mux.Handle("/api/v1/feedback.json", NewHandlerJSON(mailer))
	mux.Handle("POST "+formEndpoint, NewFormHandler(formEndpoint, mailer))
	mux.Handle("/", NewIndexHandler(formEndpoint))

	fmt.Printf(
		`Listening at http://%[1]s/

    Test URL Form Submission with localization:
      curl --header "Accept-Language: ru-RU" -v "http://%[1]s/api/v1/feedback.json?name=TestName&email=t@gmail.com&message=tryIt"
`,
		l.Addr(),
	)

	localizerContextMiddleware := acceptlanguage.New(bundle)
	http.Serve(l, localizerContextMiddleware(mux))
}
