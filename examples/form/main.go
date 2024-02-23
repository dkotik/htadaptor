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
	"github.com/dkotik/htadaptor/examples/form/feedback"
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
	localizerContextMiddleware := acceptlanguage.New(bundle)

	mux := http.NewServeMux()
	mux.Handle("/test/form", localizerContextMiddleware(htadaptor.Must(
		feedback.New(feedback.Sender(
			func(ctx context.Context, r *feedback.Request) error {
				// in real application replace with adaptor to mail provider
				slog.Default().InfoContext(
					ctx,
					"received a feedback request",
					slog.Any("request", r),
				)
				return nil
			},
		), htadaptor.WithQueryValues(
			// alternative to form body
			"name",
			"email",
			"phone",
			"message",
		)),
	)))

	fmt.Printf(
		`Listening at http://%[1]s/

    Test URL Form Submission with localization:
      curl --header "Accept-Language: ru-RU" -v "http://%[1]s/test/form?name=TestName&email=t@gmail.com&message=tryIt"
`,
		l.Addr(),
	)

	http.Serve(l, mux)
}
