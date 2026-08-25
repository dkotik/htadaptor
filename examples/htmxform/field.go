package main

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"regexp"

	"github.com/dkotik/htadaptor"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type FieldValidator func(string, string) (*i18n.LocalizeConfig, bool)

type inputValidation struct {
	Name  string
	Value string
}

func newValidationError(ctx context.Context, msg *i18n.LocalizeConfig) error {
	if msg == nil {
		return nil
	}
	lc, ok := htadaptor.LocalizerFromContext(ctx)
	if !ok {
		return errors.New("context localizer not found")
	}
	verr, err := lc.Localize(msg)
	if err != nil {
		return err
	}
	return errors.New(verr)
}

func (i *inputValidation) Validate(ctx context.Context) error {
	switch i.Name {
	case "message":
		return newValidationError(ctx, validateMessage(i.Value))
	case "name":
		return newValidationError(ctx, validateName(i.Value))
	case "email":
		return newValidationError(ctx, validateEmail(i.Value))
	case "phone":
		return newValidationError(ctx, validatePhone(i.Value))
	}
	return nil
}

func NewValidator(
	a htadaptor.Adaptor,
	fv FieldValidator,
	readLimit int64,
	template *template.Template,
) (http.Handler, error) {
	if fv == nil {
		return nil, errors.New("nil validator")
	}
	if template == nil {
		return nil, errors.New("nil template")
	}
	if readLimit < 1 {
		return nil, errors.New("read limit must be non-negative")
	}
	return a.AdaptFunc(
		func(context.Context, *inputValidation) (string, error) {
			return "", nil
		},
		htadaptor.WithEncoder(htadaptor.EncoderFunc(
			func(w http.ResponseWriter, r *http.Request, i int, a any) error {
				// empty response means everything is fine
				return nil
			},
		)),
		htadaptor.WithErrorHandler(
			htadaptor.ErrorHandlerFunc(
				func(w http.ResponseWriter, r *http.Request, err error) error {
					w.Header().Set("content-type", "text/html")
					w.WriteHeader(http.StatusUnprocessableEntity)
					return template.Execute(w, err.Error())
				},
			),
		),
	)
}

func validateName(v string) *i18n.LocalizeConfig {
	if v == "" {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorNameRequired",
				Other: "Name is required.",
			},
		}
	} else if len(v) < 4 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorNameTooShort",
				Other: "Name must be at least 4 characters.",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorNameTooLong",
				Other: "Name must be less than 64 characters.",
			},
		}
	}
	return nil
}

func validatePhone(v string) *i18n.LocalizeConfig {
	if v == "" {
		// all good
	} else if len(v) < 7 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorPhoneTooShort",
				Other: "Phone number must be at least 7 characters.",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorPhoneTooLong",
				Other: "Phone number must be less than 64 characters.",
			},
		}
	} else if !regexp.MustCompile(`^\+?[0-9\-\(\) ]{7,63}$`).MatchString(v) {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorPhoneInvalid",
				Other: "Phone number is format invalid.",
			},
		}
	}
	return nil
}

var reValidEmailAddressW3C = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func validateEmail(v string) *i18n.LocalizeConfig {
	if v == "" {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorEmailRequired",
				Other: "Email is required.",
			},
		}
	} else if len(v) < 4 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorEmailTooShort",
				Other: "Email must be at least 4 characters.",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorEmailTooLong",
				Other: "Email must be less than 64 characters.",
			},
		}
	} else if !reValidEmailAddressW3C.MatchString(v) {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorEmailInvalid",
				Other: "Email is format invalid.",
			},
		}
	}
	return nil
}

func validateMessage(v string) *i18n.LocalizeConfig {
	if v == "" {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorMessageRequired",
				Other: "Message is required.",
			},
		}
	} else if len(v) < 7 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorMessageTooShort",
				Other: "Message must be at least 7 characters.",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorMessageTooLong",
				Other: "Message must be less than 64 characters.",
			},
		}
	}
	return nil
}
