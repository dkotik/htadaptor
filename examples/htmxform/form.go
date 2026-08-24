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

func (f ContactForm) WithSchema() ContactForm {
	f.Name.Name = "name"
	f.Name.Type = "text"
	f.Name.IsRequired = true

	f.Phone.Name = "phone"
	f.Phone.Type = "text"

	f.Email.Name = "email"
	f.Email.Type = "email"
	f.Email.IsRequired = true

	f.Message.Name = "message"
	f.Message.Type = "textarea"
	f.Message.IsRequired = true
	return f
}

func (f ContactForm) WithLabels(lc *i18n.Localizer) (_ ContactForm, err error) {
	f.Title, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Contact",
			Other: "Contact",
		},
	})
	if err != nil {
		return f, err
	}
	f.Name.Label, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelName",
			Other: "Name",
		},
	})
	if err != nil {
		return f, err
	}
	f.Phone.Label, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelPhone",
			Other: "Phone",
		},
	})
	if err != nil {
		return f, err
	}
	f.Email.Label, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelEmail",
			Other: "Email",
		},
	})
	if err != nil {
		return f, err
	}
	f.Message.Label, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelMessage",
			Other: "Message",
		},
	})
	if err != nil {
		return f, err
	}
	return f, nil
}

func (f ContactForm) send() *i18n.LocalizeConfig {
	return &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LabelSend",
			Other: "Send",
		},
	}
}

func (f ContactForm) Get(ctx context.Context) (_ ContactForm, err error) {
	lc, _ := htadaptor.LocalizerFromContext(ctx)
	if lc == nil {
		return f, errors.New("nil context localizer")
	}
	f, err = f.WithLabels(lc)
	if err != nil {
		return f, err
	}
	f.Send, err = lc.Localize(f.send())
	f = f.WithSchema()
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
	Sender  Sender
	IsValid bool
}

func (f PostContactForm) Post(ctx context.Context, l *Letter) (_ PostContactForm, err error) {
	lc, _ := htadaptor.LocalizerFromContext(ctx)
	if lc == nil {
		return f, errors.New("nil context localizer")
	}
	var verr *i18n.LocalizeConfig
	f.IsValid = true

	if verr = validateName(l.Name); verr != nil {
		f.IsValid = false
		f.ContactForm.Name.Error, err = lc.Localize(verr)
		if err != nil {
			return f, err
		}
	}
	f.ContactForm.Name.Value = l.Name

	if verr = validatePhone(l.Phone); verr != nil {
		f.IsValid = false
		f.ContactForm.Phone.Error, err = lc.Localize(verr)
		if err != nil {
			return f, err
		}
	}
	f.ContactForm.Phone.Value = l.Phone

	if verr = validateEmail(l.Email); verr != nil {
		f.IsValid = false
		f.ContactForm.Email.Error, err = lc.Localize(verr)
		if err != nil {
			return f, err
		}
	}
	f.ContactForm.Email.Value = l.Email

	if verr = validateMessage(l.Message); verr != nil {
		f.IsValid = false
		f.ContactForm.Message.Error, err = lc.Localize(verr)
		if err != nil {
			return f, err
		}
	}
	f.ContactForm.Message.Value = l.Message

	if f.IsValid {
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
			// TODO: handle send error as a form error
			return f, err
		}
	}

	f.ContactForm, err = f.ContactForm.WithLabels(lc)
	if err != nil {
		return f, err
	}
	f.Send, err = lc.Localize(f.send())
	f.ContactForm = f.ContactForm.WithSchema()
	if err != nil {
		return f, err
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
	postHX, err := adaptor.AdaptFunc(
		func(ctx context.Context, l *Letter) (PostContactForm, error) {
			return PostContactForm{Sender: s}.Post(ctx, l)
		},
		htadaptor.WithTemplate(tmpl.Lookup("form")),
	)
	if err != nil {
		return nil, err
	}

	validator, err := NewValidator(
		adaptor,
		FieldValidator{
			Name:      "message",
			Validator: validateMessage,
		},
		FieldValidator{
			Name:      "email",
			Validator: validateEmail,
		},
		FieldValidator{
			Name:      "name",
			Validator: validateName,
		},
		FieldValidator{
			Name:      "phone",
			Validator: validatePhone,
		},
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
