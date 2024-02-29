package feedback

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func AddRussian(b *i18n.Bundle) error {
	return b.AddMessages(language.Russian, []*i18n.Message{
		{ID: "PageTitle", Other: "Обратная Связь"},
		{ID: "Name", Other: "Ваше Имя"},
		{ID: "Phone", Other: "Номер Телефона"},
		{ID: "Email", Other: "Адрес Электронной Почты"},
		{ID: "EmailFormatError", Other: "Ложный адрес электронной почты."},
		{ID: "Message", Other: "Сообщение"},
		{ID: "Required", Other: "Пожалуйста заполните {{.Field}}."},
		{ID: "Send", Other: "Send"},
		{ID: "Sent", Other: "Спасибо! Мы вскоре с вами свяжемся."},
		{ID: "Error", Other: "Невозможно отправить: {{.Error}}."},
	}...)
}
