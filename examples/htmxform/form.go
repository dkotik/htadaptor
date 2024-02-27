package main

import (
	"context"
	_ "embed" // for templates
	"errors"
	"html/template"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/examples/htmxform/feedback"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed form.htmx
var htmxTemplates string

var templates = template.Must(template.New("htmx").Parse(htmxTemplates))

type formRequest struct {
	feedback.Request // embed request to defer validation
}

func (r *formRequest) Validate(ctx context.Context) error {
	// do nothing, because feedback.Request validation
	// will be performed inside the handler function
	return nil
}

type formResponse struct {
	feedback.Request // embed request to inject form values into form
	NameLabel        string
	PhoneLabel       string
	EmailLabel       string
	MessageLabel     string
	SendLabel        string
	Success          string
	Error            error
}

func newFormResponse(l *i18n.Localizer) formResponse {
	return formResponse{
		NameLabel: l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: feedback.MsgName,
		}),
		PhoneLabel: l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: feedback.MsgPhone,
		}),
		EmailLabel: l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: feedback.MsgEmail,
		}),
		MessageLabel: l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: feedback.MsgMessage,
		}),
		SendLabel: l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: feedback.MsgSend,
		}),
	}
}

func NewFormHandler(sender feedback.Sender) http.Handler {
	if sender == nil {
		panic("cannot use a <nil> sender")
	}

	return htadaptor.Must(htadaptor.NewUnaryFuncAdaptor(
		func(ctx context.Context, r *formRequest) (*formResponse, error) {
			// localizer is passed through context using
			// acceptlanguage middleware all the same
			l, ok := htadaptor.LocalizerFromContext(ctx)
			if !ok {
				return nil, errors.New("there is no localizer in context")
			}
			f := newFormResponse(l)
			f.Request = r.Request

			// validation fed into template instead of
			// responding with 422 to display HTMX cleanly
			if err := r.Request.Validate(ctx); err != nil {
				f.Error = err
			} else if err = sender(ctx, &r.Request); err != nil {
				f.Error = err
			} else {
				f.Success = l.MustLocalize(&i18n.LocalizeConfig{
					MessageID: feedback.MsgSent,
				})
			}
			return &f, nil
		},
		htadaptor.WithTemplate(templates.Lookup("form")),
	))
}
