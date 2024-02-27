package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/examples/htmxform/feedback"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type page struct {
	Title        string
	formResponse // for labels
}

func NewIndexHandler() http.Handler {
	return htadaptor.Must(htadaptor.NewNullaryFuncAdaptor(
		func(ctx context.Context) (*page, error) {
			// localizer is passed through context using
			// acceptlanguage middleware all the same
			l, ok := htadaptor.LocalizerFromContext(ctx)
			if !ok {
				return nil, errors.New("there is no localizer in context")
			}
			return &page{
				Title: l.MustLocalize(&i18n.LocalizeConfig{
					MessageID: feedback.MsgSend,
				}),
				formResponse: newFormResponse(l),
			}, nil
		},
		htadaptor.WithTemplate(templates.Lookup("page")),
	))
}
