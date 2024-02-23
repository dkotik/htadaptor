/*
Package feedback is a standard contact form.
*/
package feedback

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dkotik/htadaptor"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var reValidEmail = regexp.MustCompile(`^[^\@]+\@[^\@]+\.\w+$`)

type Sender func(context.Context, *Request) error

func New(send Sender, withOptions ...htadaptor.Option) (http.Handler, error) {
	if send == nil {
		return nil, errors.New("cannot use a <nil> feedback sender")
	}
	return htadaptor.NewUnaryFuncAdaptor(
		func(ctx context.Context, r *Request) (p *Response, err error) {
			if err = send(ctx, r); err != nil {
				return nil, fmt.Errorf(htadaptor.LocalizerFromContext(ctx).MustLocalize(&i18n.LocalizeConfig{
					MessageID: MsgError,
					TemplateData: map[string]any{
						"Error": "%w",
					},
				}), err)
			}
			return &Response{
				Message: htadaptor.LocalizerFromContext(ctx).MustLocalize(
					&i18n.LocalizeConfig{
						MessageID: MsgSent,
					},
				),
			}, nil
		},
		withOptions...,
	)
}

type Request struct {
	Name    string
	Phone   string
	Email   string
	Message string
}

func (r *Request) Validate(ctx context.Context) error {
	if len(r.Name) < 4 {
		l := htadaptor.LocalizerFromContext(ctx)
		return errors.New(l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: MsgRequired,
			TemplateData: map[string]any{
				"Field": strings.ToLower(l.MustLocalize(&i18n.LocalizeConfig{
					MessageID: MsgName,
				})),
			},
		}))
	}
	if len(r.Email) < 4 {
		l := htadaptor.LocalizerFromContext(ctx)
		return errors.New(l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: MsgRequired,
			TemplateData: map[string]any{
				"Field": strings.ToLower(l.MustLocalize(&i18n.LocalizeConfig{
					MessageID: MsgEmail,
				})),
			},
		}))
	}
	if !reValidEmail.MatchString(r.Email) {
		l := htadaptor.LocalizerFromContext(ctx)
		return errors.New(l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: MsgEmailError,
		}))
	}
	if len(r.Message) < 4 {
		l := htadaptor.LocalizerFromContext(ctx)
		return errors.New(l.MustLocalize(&i18n.LocalizeConfig{
			MessageID: MsgRequired,
			TemplateData: map[string]any{
				"Field": strings.ToLower(l.MustLocalize(&i18n.LocalizeConfig{
					MessageID: MsgMessage,
				})),
			},
		}))
	}
	return nil
}

type Response struct {
	Message string `json:"message"` // to lower case
}
