package main

import (
	"net/http"

	"github.com/dkotik/htadaptor"
	"github.com/dkotik/htadaptor/examples/htmxform/feedback"
)

func NewHandlerJSON(sender feedback.Sender) http.Handler {
	if sender == nil {
		panic("cannot use a <nil> sender")
	}

	return htadaptor.Must(
		feedback.New(sender, htadaptor.WithQueryValues(
			// alternative query values to form body for testing
			"name",
			"email",
			"phone",
			"message",
		)),
	)
}
