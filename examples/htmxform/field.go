package main

import (
	"context"
	"errors"
	"html/template"
	"io"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/extract"
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
	return a.AdaptStringFunc(
		func(ctx context.Context, msg string) (string, error) {
			return msg, nil // pass through the extractor value
		},
		extract.StringValueExtractorFunc(
			func(r *http.Request) (_ string, err error) {
				type limitReadCloser struct {
					io.Reader
					io.Closer
				}

				// enforce the upper limit on the request body
				r.Body = limitReadCloser{
					Reader: io.LimitReader(r.Body, 10*1024),
					Closer: r.Body,
				}
				if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
					if err = r.ParseForm(); err != nil {
						return "", err
					}
				} else if err = r.ParseMultipartForm(10 * 1024); err != nil {
					return "", err
				}

				for key, values := range r.Form {
					for _, fv := range fvs {
						if fv.Name == key {
							verr := fv.Validator(values[0])
							if verr == nil {
								return "", nil
							}
							lc, ok := htadaptor.LocalizerFromContext(r.Context())
							if !ok {
								return "", errors.New("context localizer not found")
							}
							return lc.Localize(verr)
						}
					}
				}
				return "", errors.New("absent field value")
			},
		),
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
