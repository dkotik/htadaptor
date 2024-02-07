/*
Package extractor provides a standard set of most common [http.Request]
value extractors which populate fields of a decoded generic request struct.

The most of the extractors target [url.Values] because it preserves
duplicate fields. This gives the flexibility to a [htadaptor.Decoder]
to choose how to handle the duplicates.
*/
package extractor

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// RequestValueExtractor pulls [url.Values] from an [http.Request]
// in order to provide a [Decoder] with values to populate
// domain request struct with.
type RequestValueExtractor interface {
	ExtractRequestValue(url.Values, *http.Request) error
}

// RequestValueExtractorFunc provides a wrapper for functional
// implementation of a [RequestValueExtractor].
type RequestValueExtractorFunc func(url.Values, *http.Request) error

// ExtractRequestValue satisfies [RequestValueExtractor] interface
// for [RequestValueExtractorFunc].
func (f RequestValueExtractorFunc) ExtractRequestValue(vs url.Values, r *http.Request) error {
	return f(vs, r)
}

// Join unites several extractors into one.
// Returns <nil> if no extractors are given.
func Join(exs ...RequestValueExtractor) RequestValueExtractor {
	switch len(exs) {
	case 0:
		return nil
	case 1:
		return exs[0]
	default:
		return RequestValueExtractorFunc(
			func(vs url.Values, r *http.Request) (err error) {
				for _, extractor := range exs {
					if err = extractor.ExtractRequestValue(vs, r); err != nil {
						return err
					}
				}
				return nil
			},
		)
	}
}

func uniqueNonEmptyValueNames(names []string) error {
	if len(names) == 0 {
		return errors.New("provide at least one value name")
	}
	found := make(map[string]struct{})
	for _, name := range names {
		if name == "" {
			return errors.New("cannot use an empty value name")
		}
		if _, ok := found[name]; ok {
			return fmt.Errorf("value name %q occurs more than once", name)
		}
		found[name] = struct{}{}
	}
	return nil
}
