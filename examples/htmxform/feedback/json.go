package feedback

import (
	"net/http"
  "errors"
  "context"

	"github.com/dkotik/htadaptor"
)

func NewJSON(sender Sender, withOptions ...htadaptor.Option) (http.Handler, error) {
  if sender == nil {
		return nil, errors.New("cannot use a <nil> feedback sender")
	}
	return htadaptor.NewVoidFuncAdaptor(
  		func(ctx context.Context, r *Letter) (err error) {
  			return sender(ctx, r)
  		},
  		withOptions...,
  	)
}
