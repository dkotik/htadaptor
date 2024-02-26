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
	formResponse // for labels
}

func (p *page) Title() (string, error) {
	return p.formResponse.Localizer.Localize(&i18n.LocalizeConfig{
		MessageID: feedback.MsgSend,
	})
}

func NewIndexHandler(formTarget string) http.Handler {
	return htadaptor.Must(htadaptor.NewNullaryFuncAdaptor(
		func(ctx context.Context) (*page, error) {
			// localizer is passed through context using
			// acceptlanguage middleware all the same
			l := htadaptor.LocalizerFromContext(ctx)
			if l == nil {
				return nil, errors.New("there is no localizer in context")
			}
			return &page{
				formResponse: formResponse{
					FormTarget: formTarget,
					Localizer:  l,
				},
			}, nil
		},
		htadaptor.WithTemplate(templates.Lookup("page")),
	))
}
