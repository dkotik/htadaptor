package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type FormField struct {
	Label      string
	Name       string
	Type       string
	Value      string
	Help       string
	Error      string
	IsRequired bool
}

type FieldValidator struct {
	Name      string
	Validator func(string) *i18n.LocalizeConfig
}

type fieldValue struct {
	Name  string
	Value string
}

func (v *fieldValue) Validate(context.Context) error {
	return nil
}

func NewValidator(
	a htadaptor.Adaptor,
	fvs ...FieldValidator,
) (http.Handler, error) {
	for _, fv := range fvs {
		if fv.Validator == nil {
			return nil, errors.New("nil validator")
		}
	}
	return a.AdaptFunc(
		func(ctx context.Context, field *fieldValue) (string, error) {
			for _, fv := range fvs {
				if fv.Name == field.Name {
					lc, ok := htadaptor.LocalizerFromContext(ctx)
					if !ok {
						return "", errors.New("context localizer not found")
					}
					verr := fv.Validator(field.Value)
					if verr == nil {
						return "", nil
					}
					return lc.Localize(verr)
				}
			}
			return "", fmt.Errorf("unknown field: %s", field.Name)
		},
		htadaptor.WithDecoder(htadaptor.DecoderFunc(
			func(v any, r *http.Request) (err error) {
				if err = r.ParseForm(); err != nil {
					return err
				}
				fv, ok := v.(*fieldValue)
				if !ok {
					return errors.New("not a field value type")
				}
				for key, values := range r.Form {
					// take the first value
					fv.Name = key
					fv.Value = values[0]
					return nil
				}
				return errors.New("no field value")
			},
		)),
		htadaptor.WithTemplate(
			template.Must(
				template.New("").Parse(`
					{{- with . -}}
					  <p class="help is-danger">{{ . }}</p>
					{{- end -}}
				`),
			),
		),
		htadaptor.WithErrorHandler(htadaptor.ErrorHandlerFunc(
			func(w http.ResponseWriter, r *http.Request, err error) error {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return nil
			},
		)),
	)
}

func validateName(v string) *i18n.LocalizeConfig {
	if v == "" {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "NameRequired",
				Other: "Name is required",
			},
		}
	} else if len(v) < 4 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "NameTooShort",
				Other: "Name must be at least 4 characters",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "NameTooLong",
				Other: "Name must be less than 64 characters",
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
				ID:    "PhoneTooShort",
				Other: "Phone number must be at least 7 characters",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PhoneTooLong",
				Other: "Phone number must be less than 64 characters",
			},
		}
	}
	return nil
}

func validateEmail(v string) *i18n.LocalizeConfig {
	if v == "" {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EmailRequired",
				Other: "Email is required",
			},
		}
	} else if len(v) < 4 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EmailTooShort",
				Other: "Email must be at least 4 characters",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EmailTooLong",
				Other: "Email must be less than 64 characters",
			},
		}
	}
	return nil
}

func validateMessage(v string) *i18n.LocalizeConfig {
	if v == "" {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "MessageRequired",
				Other: "Message is required.",
			},
		}
	} else if len(v) < 7 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "MessageTooShort",
				Other: "Message must be at least 7 characters",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "MessageTooLong",
				Other: "Message must be less than 64 characters",
			},
		}
	}
	return nil
}
