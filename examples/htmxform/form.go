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
	FormTarget       string
	feedback.Request // embed request to inject form values into form
	Success          string
	Error            error
	Localizer        *i18n.Localizer
}

func NewFormHandler(formTarget string, sender feedback.Sender) http.Handler {
	if sender == nil {
		panic("cannot use a <nil> sender")
	}

	return htadaptor.Must(htadaptor.NewUnaryFuncAdaptor(
		func(ctx context.Context, r *formRequest) (*formResponse, error) {
			// localizer is passed through context using
			// acceptlanguage middleware all the same
			l := htadaptor.LocalizerFromContext(ctx)
			if l == nil {
				return nil, errors.New("there is no localizer in context")
			}
			f := &formResponse{
				FormTarget: formTarget,
				Request:    r.Request,
				Localizer:  l,
			}

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
			return f, nil
		},
		htadaptor.WithTemplate(templates.Lookup("form")),
	))
}

func (r *formResponse) NameLabel() (string, error) {
	return r.Localizer.Localize(&i18n.LocalizeConfig{
		MessageID: feedback.MsgName,
	})
}

func (r *formResponse) EmailLabel() (string, error) {
	return r.Localizer.Localize(&i18n.LocalizeConfig{
		MessageID: feedback.MsgEmail,
	})
}

func (r *formResponse) PhoneLabel() (string, error) {
	return r.Localizer.Localize(&i18n.LocalizeConfig{
		MessageID: feedback.MsgPhone,
	})
}

func (r *formResponse) MessageLabel() (string, error) {
	return r.Localizer.Localize(&i18n.LocalizeConfig{
		MessageID: feedback.MsgMessage,
	})
}

func (r *formResponse) SendLabel() (string, error) {
	return r.Localizer.Localize(&i18n.LocalizeConfig{
		MessageID: feedback.MsgSend,
	})
}
