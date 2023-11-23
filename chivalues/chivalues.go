/*
Package chivalues provides value extractor for Chi router.
*/
package chivalues

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type ChiRequestValueExtractor struct {
	names []string
}

func New(names ...string) (*ChiRequestValueExtractor, error) {
	if len(names) == 0 {
		return nil, errors.New("provide at least one URL query parameter name")
	}
	found := make(map[string]struct{})
	for _, name := range names {
		if name == "" {
			return nil, errors.New("cannot use an empty URL query parameter name")
		}
		if _, ok := found[name]; ok {
			return nil, fmt.Errorf("URL query parameter %q is listed more than once", name)
		}
		found[name] = struct{}{}
	}
	return &ChiRequestValueExtractor{names: names}, nil
}

func (c *ChiRequestValueExtractor) ExtractRequestValue(r *http.Request) (url.Values, error) {
	result := make(url.Values)
	rctx := chi.RouteContext(r.Context())
	for i, name := range rctx.URLParams.Keys {
		for _, desired := range c.names {
			if name == desired {
				result[name] = []string{rctx.URLParams.Values[i]}
				break
			}
		}
	}
	return result, nil
}
