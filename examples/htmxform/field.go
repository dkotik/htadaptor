package main

import (
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

func (f FormField) IsValid() bool {
	if f.Error != "" {
		return false
	}
	return true
}

type fieldValidator struct {
	Input    *template.Template
	TextArea *template.Template
}

func NewFieldValidator(
	input *template.Template,
	textarea *template.Template,
) (http.Handler, error) {
	if input == nil {
		return nil, errors.New("input template is nil")
	}
	if textarea == nil {
		return nil, errors.New("textarea template is nil")
	}
	h := fieldValidator{
		Input:    input,
		TextArea: textarea,
	}
	// Limit the total request body to 10 Kilobytes
	return http.MaxBytesHandler(h, 10*1024), nil
}

func (v fieldValidator) GetFieldModel(r *http.Request) (_ any, _ *template.Template, err error) {
	lc, _ := htadaptor.LocalizerFromContext(r.Context())
	if lc == nil {
		return nil, nil, errors.New("nil context localizer")
	}
	if err = r.ParseForm(); err != nil {
		return nil, nil, err
	}
	for key, values := range r.Form {
		switch key {
		case "name":
			name, err := NewNameFieldWithValue(lc, values[0])
			if err != nil {
				return nil, nil, err
			}
			return name, v.Input, nil
		case "phone":
			phone, err := NewPhoneFieldWithValue(lc, values[0])
			if err != nil {
				return nil, nil, err
			}
			return phone, v.Input, nil
		case "email":
			email, err := NewEmailFieldWithValue(lc, values[0])
			if err != nil {
				return nil, nil, err
			}
			return email, v.Input, nil
		case "message":
			message, err := NewMessageFieldWithValue(lc, values[0])
			if err != nil {
				return nil, nil, err
			}
			return message, v.TextArea, nil
		default:
			return nil, nil, fmt.Errorf("unknown field: %s", key)
		}
	}
	return nil, nil, errors.New("zero form fields")
}

func (v fieldValidator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	model, t, err := v.GetFieldModel(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.WriteHeader(http.StatusOK)
	if err = t.Execute(w, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewNameField(lc *i18n.Localizer) (f FormField, err error) {
	f.Name = "name"
	f.Type = "text"
	f.IsRequired = true
	f.Label, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Name",
			Other: "Your Name",
		},
	})
	return f, err
}

func NewNameFieldWithValue(lc *i18n.Localizer, v string) (f FormField, err error) {
	f, err = NewNameField(lc)
	if err != nil {
		return f, err
	}
	f.Value = v
	if v == "" {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "NameRequired",
				Other: "Name is required",
			},
		})
		if err != nil {
			return f, err
		}
	} else if len(v) < 4 {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "NameTooShort",
				Other: "Name must be at least 4 characters",
			},
		})
		if err != nil {
			return f, err
		}
	} else if len(v) >= 64 {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "NameTooLong",
				Other: "Name must be less than 64 characters",
			},
		})
		if err != nil {
			return f, err
		}
	}
	return f, nil
}

func NewPhoneField(lc *i18n.Localizer) (f FormField, err error) {
	f.Name = "phone"
	f.Type = "phone"
	f.Label, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Phone",
			Other: "Phone Number",
		},
	})
	return f, err
}

func NewPhoneFieldWithValue(lc *i18n.Localizer, v string) (f FormField, err error) {
	f, err = NewPhoneField(lc)
	if err != nil {
		return f, err
	}
	f.Value = v
	if v == "" {
		// all good
	} else if len(v) < 7 {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PhoneTooShort",
				Other: "Phone number must be at least 7 characters",
			},
		})
		if err != nil {
			return f, err
		}
	} else if len(v) >= 64 {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PhoneTooLong",
				Other: "Phone number must be less than 64 characters",
			},
		})
		if err != nil {
			return f, err
		}
	}
	return f, nil
}

func NewEmailField(lc *i18n.Localizer) (f FormField, err error) {
	f.Name = "email"
	f.Type = "email"
	f.IsRequired = true
	f.Label, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Email",
			Other: "E-mail Address",
		},
	})
	return f, err
}

func NewEmailFieldWithValue(lc *i18n.Localizer, v string) (f FormField, err error) {
	f, err = NewEmailField(lc)
	if err != nil {
		return f, err
	}
	f.Value = v
	if v == "" {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EmailRequired",
				Other: "Email is required",
			},
		})
		if err != nil {
			return f, err
		}
	} else if len(v) < 4 {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EmailTooShort",
				Other: "Email must be at least 4 characters",
			},
		})
		if err != nil {
			return f, err
		}
	} else if len(v) >= 64 {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EmailTooLong",
				Other: "Email must be less than 64 characters",
			},
		})
		if err != nil {
			return f, err
		}
	}
	return f, nil
}

func NewMessageField(lc *i18n.Localizer) (f FormField, err error) {
	f.Name = "message"
	f.Type = "text"
	f.Label, err = lc.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Message",
			Other: "Message",
		},
	})
	return f, err
}

func NewMessageFieldWithValue(lc *i18n.Localizer, v string) (f FormField, err error) {
	f, err = NewMessageField(lc)
	if err != nil {
		return f, err
	}
	f.Value = v
	if v == "" {
		// all good
	} else if len(v) < 7 {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "MessageTooShort",
				Other: "Message number must be at least 7 characters",
			},
		})
		if err != nil {
			return f, err
		}
	} else if len(v) >= 64 {
		f.Error, err = lc.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "MessageTooLong",
				Other: "Message number must be less than 64 characters",
			},
		})
		if err != nil {
			return f, err
		}
	}
	return f, nil
}
