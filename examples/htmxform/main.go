/*
Package main demonstrates an implementation of a standard
feedback form with validation and localization.
*/
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/examples/htmxform/feedback"
	"github.com/dkotik/htadaptor/middleware/acceptlanguage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// in real application replace with adaptor to mail provider
var mailer = feedback.Sender(
	func(ctx context.Context, r *feedback.Letter) error {
		slog.Default().InfoContext(
			ctx,
			"received a feedback request",
			slog.Any("request", r),
		)
		return nil
	},
)

func main() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	bundle := i18n.NewBundle(language.English)
	feedback.AddRussian(bundle)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/feedback.json", htadaptor.Must(feedback.NewJSON(
		mailer,
		htadaptor.WithQueryValues(
			// alternative query values to form body for testing
			"name",
			"email",
			"phone",
			"message",
		),
	)))
	getForm, postForm, err := feedback.New(mailer)
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /{$}", getForm)
	mux.Handle("POST /{$}", postForm)

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
