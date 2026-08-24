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
	Title   string
	Name    FormField
	Phone   FormField
	Email   FormField
	Message FormField
	Send    string
	Sent    string
}

func (f ContactForm) title() *i18n.LocalizeConfig {
	return &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Contact",
			Other: "Contact",
		},
	}
}

func (f ContactForm) send() *i18n.LocalizeConfig {
	return &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Send",
			Other: "Send",
		},
	}
}

func (f ContactForm) Get(ctx context.Context) (_ ContactForm, err error) {
	lc, _ := htadaptor.LocalizerFromContext(ctx)
	if lc == nil {
		return f, errors.New("nil context localizer")
	}
	f.Title, err = lc.Localize(f.title())
	if err != nil {
		return f, err
	}
	f.Name, err = NewNameField(lc)
	if err != nil {
		return f, err
	}
	f.Phone, err = NewPhoneField(lc)
	if err != nil {
		return f, err
	}
	f.Email, err = NewEmailField(lc)
	if err != nil {
		return f, err
	}
	f.Message, err = NewMessageField(lc)
	if err != nil {
		return f, err
	}
	f.Send, err = lc.Localize(f.send())
	if err != nil {
		return f, err
	}
	return f, nil
}

type Sender interface {
	Send(ctx context.Context, l *Letter) error
}

type PostContactForm struct {
	ContactForm
	Sender Sender
}

func (f PostContactForm) IsValid() bool {
	return f.ContactForm.Name.Error == "" &&
		f.ContactForm.Phone.Error == "" &&
		f.ContactForm.Email.Error == "" &&
		f.ContactForm.Message.Error == ""
}

func (f PostContactForm) Post(ctx context.Context, l *Letter) (_ PostContactForm, err error) {
	lc, _ := htadaptor.LocalizerFromContext(ctx)
	if lc == nil {
		return f, errors.New("nil context localizer")
	}
	f.Title, err = lc.Localize(f.title())
	if err != nil {
		return f, err
	}
	f.Name, err = NewNameFieldWithValue(lc, l.Name)
	if err != nil {
		return f, err
	}
	f.Phone, err = NewPhoneFieldWithValue(lc, l.Phone)
	if err != nil {
		return f, err
	}
	f.Email, err = NewEmailFieldWithValue(lc, l.Email)
	if err != nil {
		return f, err
	}
	f.Message, err = NewMessageFieldWithValue(lc, l.Message)
	if err != nil {
		return f, err
	}
	f.Send, err = lc.Localize(f.send())
	if err != nil {
		return f, err
	}
	if f.IsValid() {
		f.Sent, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Sent",
				Other: "Thank you! We will follow up with you soon.",
			},
		})
		if err != nil {
			return f, err
		}
		err = f.Sender.Send(ctx, l)
		if err != nil {
			f.Sent = ""
			return f, err
		}
	}
	return f, nil
}

func NewContactForm(s Sender) (http.Handler, error) {
	tmpl, err := template.New("").Parse(htmx)
	if err != nil {
		return nil, err
	}
	adaptor := htadaptor.New()
	get, err := adaptor.AdaptNullaryFunc(
		func(ctx context.Context) (ContactForm, error) {
			return ContactForm{}.Get(ctx)
		},
		htadaptor.WithTemplate(tmpl.Lookup("page")),
	)
	if err != nil {
		return nil, err
	}

	post, err := adaptor.AdaptFunc(
		func(ctx context.Context, l *Letter) (PostContactForm, error) {
			return PostContactForm{Sender: s}.Post(ctx, l)
		},
		htadaptor.WithTemplate(tmpl.Lookup("page")),
	)
	if err != nil {
		return nil, err
	}

	validator, err := NewFieldValidator(
		tmpl.Lookup("input"),
		tmpl.Lookup("textarea"),
	)
	if err != nil {
		return nil, err
	}
	return htadaptor.NewMethodMux(&htadaptor.MethodSwitch{
		Get:   get,
		Post:  post,
		Patch: validator,
	}), nil
}
