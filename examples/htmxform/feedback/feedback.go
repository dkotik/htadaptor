/*
Package feedback is a standard contact form.
*/
package feedback

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/dkotik/htadaptor"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var reValidEmail = regexp.MustCompile(`^[^\@]+\@[^\@]+\.\w+$`)

type Sender func(context.Context, *Letter) error

type Letter struct {
	Name    string
	Phone   string
	Email   string
	Message string
}

func (l *Letter) Validate(ctx context.Context) error {
	locale, ok := htadaptor.LocalizerFromContext(ctx)
	if !ok {
		return errors.New("there is no localizer in context")
	}
	// separating localized validation, because HTMX handler may
	// call it directly
	return l.ValidateWithLocale(locale)
}

func newRequiredError(field string, l *i18n.Localizer) error {
	msg, err := l.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Required",
			Other: "Please provide {{.Field}}.",
		},
		TemplateData: map[string]any{
			"Field": strings.ToLower(field),
		},
	})
	if err != nil {
		return err
	}
	return errors.New(msg)
}

func (l *Letter) ValidateWithLocale(locale *i18n.Localizer) error {
	if len(l.Name) < 4 {
		field, err := l.nameLabel(locale)
		if err != nil {
			return err
		}
		return newRequiredError(field, locale)
	}
	if len(l.Email) < 4 {
		field, err := l.emailLabel(locale)
		if err != nil {
			return err
		}
		return newRequiredError(field, locale)
	}
	if !reValidEmail.MatchString(l.Email) {
		errorMessage, err := locale.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EmailFormatError",
				Other: "Invalid electronic mail address.",
			},
		})
		if err != nil {
			return err
		}
		return errors.New(errorMessage)
	}
	if len(l.Message) < 4 {
		field, err := l.messageLabel(locale)
		if err != nil {
			return err
		}
		return newRequiredError(field, locale)
	}
	return nil
}

func (l *Letter) nameLabel(locale *i18n.Localizer) (string, error) {
	return locale.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Name",
			Other: "Your Name",
		},
	})
}

func (l *Letter) phoneLabel(locale *i18n.Localizer) (string, error) {
	return locale.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Phone",
			Other: "Phone Number",
		},
	})
}

func (l *Letter) emailLabel(locale *i18n.Localizer) (string, error) {
	return locale.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Email",
			Other: "Email Address",
		},
	})
}

func (l *Letter) messageLabel(locale *i18n.Localizer) (string, error) {
	return locale.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Message",
			Other: "Message",
		},
	})
}
