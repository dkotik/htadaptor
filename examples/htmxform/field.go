package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"regexp"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/extract"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type FieldValidator func(string, string) (*i18n.LocalizeConfig, bool)

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
					Reader: io.LimitReader(r.Body, readLimit),
					Closer: r.Body,
				}
				if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
					if err = r.ParseForm(); err != nil {
						return "", err
					}
				} else if err = r.ParseMultipartForm(readLimit); err != nil {
					return "", err
				}

				for key, values := range r.Form {
					verr, ok := fv(key, values[0])
					if !ok {
						return "", fmt.Errorf("unexpected validation field: %q", key)
					}
					if verr == nil {
						return "", nil
					}
					lc, ok := htadaptor.LocalizerFromContext(r.Context())
					if !ok {
						return "", errors.New("context localizer not found")
					}
					msg, err := lc.Localize(verr)
					if err != nil {
						return "", err
					}
					return "", errors.New(msg)
				}
				return "", errors.New("validation field absent")
			},
		),
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
				Other: "Name is required",
			},
		}
	} else if len(v) < 4 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorNameTooShort",
				Other: "Name must be at least 4 characters",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorNameTooLong",
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
				ID:    "ErrorPhoneTooShort",
				Other: "Phone number must be at least 7 characters",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorPhoneTooLong",
				Other: "Phone number must be less than 64 characters",
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
				Other: "Email is required",
			},
		}
	} else if len(v) < 4 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorEmailTooShort",
				Other: "Email must be at least 4 characters",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorEmailTooLong",
				Other: "Email must be less than 64 characters",
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
				Other: "Message must be at least 7 characters",
			},
		}
	} else if len(v) >= 64 {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorMessageTooLong",
				Other: "Message must be less than 64 characters",
			},
		}
	}
	return nil
}
