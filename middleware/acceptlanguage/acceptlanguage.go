/*
Package acceptlanguage provides [htadaptor.Middleware] that injects
[language.Tag] into request [context.Context] which can
be recovered using [htadaptor.LanguageFromContext].
*/
package acceptlanguage

import (
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type localizer struct {
	next   http.Handler
	bundle *i18n.Bundle
	// globalAcceptLanguages string
}

func (l *localizer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.next.ServeHTTP(w, r.WithContext(
		htadaptor.ContextWithLocalizer(
			r.Context(),
			i18n.NewLocalizer(
				l.bundle,
				r.Header.Get("Accept-Language"),
				// l.globalAcceptLanguages, // inefficient, but cannot change API
			),
		),
	))
}

func New(b *i18n.Bundle) htadaptor.Middleware {
	// , preferred ...language.Tag
	// tags := make([]string, len(preferred))
	// for i, tag := range preferred {
	// 	tags[i] = tag.String()
	// }
	// globalAcceptLanguages := strings.Join(tags, ";")
	return func(next http.Handler) http.Handler {
		return &localizer{
			next:   next,
			bundle: b,
			// globalAcceptLanguages: globalAcceptLanguages,
		}
	}
}

type negotiator struct {
	next    http.Handler
	matcher language.Matcher
}

func (n *negotiator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	// the default language will be selected for t == nil.
	tag, _, _ := n.matcher.Match(t...)
	n.next.ServeHTTP(w, r.WithContext(
		htadaptor.ContextWithLanguage(r.Context(), tag),
	))
}

func NewNegotiator(preferred ...language.Tag) htadaptor.Middleware {
	matcher := language.NewMatcher(preferred)
	return func(next http.Handler) http.Handler {
		return &negotiator{
			next:    next,
			matcher: matcher,
		}
	}
}
