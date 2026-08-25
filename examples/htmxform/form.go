package main

import (
	"context"
	_ "embed" // for form.html template
	"errors"
	"html/template"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed form.html
var htmx string

type Letter struct {
	Name    string
	Phone   string
	Email   string
	Message string
}

func (l *Letter) Validate(ctx context.Context) error {
	// The HTMX form itself will handle validation
	// because form entry errors are not execution errors.
	// They are user prompts for correcting input.
	return nil
}

type ContactForm struct {
	Letter  Letter
	IsValid bool

	localizer    *i18n.Localizer
	nameError    *i18n.LocalizeConfig
	emailError   *i18n.LocalizeConfig
	phoneError   *i18n.LocalizeConfig
	messageError *i18n.LocalizeConfig
}

func newContactForm(ctx context.Context) (ContactForm, error) {
	lc, ok := htadaptor.LocalizerFromContext(ctx)
	if !ok {
		return ContactForm{}, errors.New("no localizer in context")
	}
	return ContactForm{
		localizer: lc,
	}, nil
}

func (f ContactForm) Title() (string, error) {
	return f.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Contact",
			Other: "Contact",
		},
	})
}

func (f ContactForm) NameLabel() (string, error) {
	return f.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelName",
			Other: "Name",
		},
	})
}

func (f ContactForm) NameError() (string, error) {
	if f.nameError == nil {
		return "", nil
	}
	return f.localizer.Localize(f.nameError)
}

func (f ContactForm) PhoneLabel() (string, error) {
	return f.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelPhone",
			Other: "Phone",
		},
	})
}

func (f ContactForm) PhoneError() (string, error) {
	if f.phoneError == nil {
		return "", nil
	}
	return f.localizer.Localize(f.phoneError)
}

func (f ContactForm) EmailLabel() (string, error) {
	return f.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelEmail",
			Other: "Email",
		},
	})
}

func (f ContactForm) EmailError() (string, error) {
	if f.emailError == nil {
		return "", nil
	}
	return f.localizer.Localize(f.emailError)
}

func (f ContactForm) MessageLabel() (string, error) {
	return f.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelMessage",
			Other: "Message",
		},
	})
}

func (f ContactForm) MessageError() (string, error) {
	if f.messageError == nil {
		return "", nil
	}
	return f.localizer.Localize(f.messageError)
}

func (f ContactForm) SendLabel() (string, error) {
	return f.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelSend",
			Other: "Send",
		},
	})
}

func (f ContactForm) SentLabel() (string, error) {
	return f.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Sent",
			Other: "Thank you! We will follow up with you soon.",
		},
	})
}

type Sender interface {
	Send(ctx context.Context, l *Letter) error
}

func NewContactForm(s Sender) (http.Handler, error) {
	tmpl, err := template.New("").Parse(htmx)
	if err != nil {
		return nil, err
	}
	adaptor := htadaptor.New()
	get, err := adaptor.AdaptNullaryFunc(
		func(ctx context.Context) (ContactForm, error) {
			return newContactForm(ctx)
		},
		htadaptor.WithTemplate(tmpl.Lookup("page")),
	)
	if err != nil {
		return nil, err
	}

	send := func(ctx context.Context, l *Letter) (ContactForm, error) {
		form, err := newContactForm(ctx)
		if err != nil {
			return form, err
		}
		form = form.Validate(l)
		if !form.IsValid {
			return form, nil
		}
		if err = s.Send(ctx, l); err != nil {
			form.IsValid = false
			form.messageError = &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "SenderFailed",
					Other: "Delivery failed: {{ .Error }}.",
				},
				TemplateData: map[string]any{
					"Error": err.Error(),
				},
			}
		}
		return form, nil
	}

	post, err := adaptor.AdaptFunc(
		send,
		htadaptor.WithTemplate(tmpl.Lookup("page")),
	)
	if err != nil {
		return nil, err
	}
	postHX, err := adaptor.AdaptFunc(
		send,
		htadaptor.WithTemplate(tmpl.Lookup("form")),
	)
	if err != nil {
		return nil, err
	}

	validator, err := NewValidator(
		adaptor,
		tmpl.Lookup("validationError"),
	)
	if err != nil {
		return nil, err
	}
	return htadaptor.NewMethodMux(&htadaptor.MethodSwitch{
		Get:   get,
		Post:  htadaptor.NewHTMXSwitch(post, postHX),
		Patch: validator,
	}), nil
}
