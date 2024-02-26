package feedback

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const (
	MsgName       = "name"
	MsgPhone      = "phone"
	MsgEmail      = "email"
	MsgEmailError = "emailError"
	MsgMessage    = "message"
	MsgRequired   = "required"
	MsgSend       = "send"
	MsgSent       = "sent"
	MsgError      = "error"
)

func LoadEnglish(b *i18n.Bundle) error {
	return b.AddMessages(language.English, []*i18n.Message{
		{ID: MsgName, Other: "Your Name"},
		{ID: MsgPhone, Other: "Phone Number"},
		{ID: MsgEmail, Other: "Email Address"},
		{ID: MsgEmailError, Other: "Invalid electronic mail address."},
		{ID: MsgMessage, Other: "Message"},
		{ID: MsgRequired, Other: "Field \"{{.Field}}\" is required."},
		{ID: MsgSend, Other: "Send"},
		{ID: MsgSent, Other: "Thank you! We will follow up with you soon."},
		{ID: MsgError, Other: "Cannot accept this message: {{.Error}}."},
	}...)
}

func LoadRussian(b *i18n.Bundle) error {
	return b.AddMessages(language.Russian, []*i18n.Message{
		{ID: MsgName, Other: "Ваше Имя"},
		{ID: MsgPhone, Other: "Номер Телефона"},
		{ID: MsgEmail, Other: "Адрес Электронной Почты"},
		{ID: MsgEmailError, Other: "Ложный адрес электронной почты."},
		{ID: MsgMessage, Other: "Сообщение"},
		{ID: MsgRequired, Other: "Поле \"{{.Field}}\" необходимо."},
		{ID: MsgSend, Other: "Send"},
		{ID: MsgSent, Other: "Спасибо! Мы вскоре с вами свяжемся."},
		{ID: MsgError, Other: "Невозможно отправить: {{.Error}}."},
	}...)
}
