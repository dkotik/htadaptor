package feedback

import (
	"context"
	_ "embed" // for templates
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed form.htmx
var htmx string

var templates = template.Must(template.New("htmx").Parse(htmx))

type formRequest struct {
	Letter // embed letter to defer validation handler function
}

func (r *formRequest) Validate(ctx context.Context) error {
	// do nothing, because Letter validation
	// will be performed inside the handler function
	// because HTMX will render the error together with
	// the rest of the response
	return nil
}

type formResponse struct {
	Letter    // embed request to inject form values into form
	Sent      bool
	Localizer *i18n.Localizer
	Error     error
}

func (f *formResponse) Title() (string, error) {
	return f.Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "PageTitle",
			Other: "Provide Feedback",
		},
	})
}

func (f *formResponse) NameLabel() (string, error) {
	return f.Letter.nameLabel(f.Localizer)
}

func (f *formResponse) PhoneLabel() (string, error) {

	return f.Letter.phoneLabel(f.Localizer)
}

func (f *formResponse) EmailLabel() (string, error) {

	return f.Letter.emailLabel(f.Localizer)
}

func (f *formResponse) MessageLabel() (string, error) {
	return f.Letter.messageLabel(f.Localizer)
}

func (f *formResponse) SendLabel() (string, error) {
	return f.Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Send",
			Other: "Send",
		},
	})
}

func (f *formResponse) Success() (string, error) {
	return f.Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Sent",
			Other: "Thank you! We will follow up with you soon.",
		},
	})
}

type formHandler struct {
	get  http.Handler
	post http.Handler
}

func (h *formHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get.ServeHTTP(w, r)
	case http.MethodPost:
		h.post.ServeHTTP(w, r)
	default:
		h.get.ServeHTTP(w, r)
	}
}

func New(sender Sender) (http.Handler, error) {
	if sender == nil {
		return nil, errors.New("cannot use a <nil> feedback sender")
	}

	get, err := htadaptor.NewNullaryFuncAdaptor(
		func(ctx context.Context) (*formResponse, error) {
			// localizer is passed through context using
			// acceptlanguage middleware all the same
			l, ok := htadaptor.LocalizerFromContext(ctx)
			if !ok {
				return nil, errors.New("there is no localizer in context")
			}
			return &formResponse{
				Localizer: l,
			}, nil
		},
		htadaptor.WithTemplate(templates.Lookup("page")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create get handler: %w", err)
	}

	post, err := htadaptor.NewUnaryFuncAdaptor(
		func(ctx context.Context, r *formRequest) (*formResponse, error) {
			// localizer is passed through context using
			// acceptlanguage middleware
			l, ok := htadaptor.LocalizerFromContext(ctx)
			if !ok {
				return nil, errors.New("there is no localizer in context")
			}
			f := &formResponse{
				Letter:    r.Letter,
				Localizer: l,
			}
			if f.Error = r.Letter.ValidateWithLocale(l); f.Error != nil {
				return f, nil
			}
			f.Error = sender(ctx, &r.Letter)
			if f.Error == nil {
				f.Sent = true
			}
			return f, nil
		},
		htadaptor.WithTemplate(templates.Lookup("form")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create post handler: %w", err)
	}

	return &formHandler{
		get:  get,
		post: post,
	}, nil
}
