package htadaptor

import (
	"context"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type contextKey struct{}

var (
	languageContextKey  = contextKey{}
	localizerContextKey = contextKey{}
)

func LanguageFromContext(ctx context.Context) language.Tag {
	tag, ok := ctx.Value(languageContextKey).(language.Tag)
	if ok {
		return tag
	}
	return language.English
}

func ContextWithLanguage(parent context.Context, t language.Tag) context.Context {
	return context.WithValue(parent, languageContextKey, t)
}

// LocalizerFromContext raises request-scoped localizer.
// Warning: localizer will be <nil> if it was not set
// using [ContextWithLocalizer].
func LocalizerFromContext(ctx context.Context) *i18n.Localizer {
	return ctx.Value(languageContextKey).(*i18n.Localizer)
}

func ContextWithLocalizer(parent context.Context, l *i18n.Localizer) context.Context {
	return context.WithValue(parent, localizerContextKey, l)
}
