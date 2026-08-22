package feedback

import (
	"context"
	"errors"
	"net/http"

	"github.com/dkotik/htadaptor"
)

func NewJSON(sender Sender, withOptions ...htadaptor.Option) (http.Handler, error) {
	if sender == nil {
		return nil, errors.New("cannot use a <nil> feedback sender")
	}
	return htadaptor.New().AdaptVoidFunc(
		func(ctx context.Context, r *Letter) (err error) {
			return sender(ctx, r)
		},
		withOptions...,
	)
}
