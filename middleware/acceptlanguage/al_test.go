package acceptlanguage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dkotik/htadaptor"
	"golang.org/x/text/language"
)

func TestAcceptLanguageHeader(t *testing.T) {
	mw := NewNegotiator(
		language.Ukrainian,
		language.Kazakh,
		language.English,
	)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tag, _ := htadaptor.LanguageFromContext(r.Context()).Base()
		if tag.String() != "kk" {
			t.Fatal("language base does not match", tag)
		}
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Language", "kk-KZ")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, request)
}
