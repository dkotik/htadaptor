/*
Package extract provides a standard set of most common [http.Request]
value extractors which populate fields of a decoded generic request struct.

The most of the extractors target [url.Values] because it preserves
duplicate fields. This gives the flexibility to a [htadaptor.Decoder]
to choose how to handle the duplicates.
*/
package extract

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// NoValueError indicates that none of the required values
// from a list of possible value names were extracted.
type NoValueError []string

func (e NoValueError) Error() string {
	switch len(e) {
	case 0:
		return "request does not include required value"
	case 1:
		return fmt.Sprintf("request requires %q value", e[0])
	default:
		b := strings.Builder{}
		b.WriteString("request requires any of ")
		for _, name := range e {
			b.WriteRune('"')
			b.WriteString(name)
			b.WriteRune('"')
			b.WriteRune(',')
			b.WriteRune(' ')
		}
		return b.String()[:b.Len()-2] + " values"
	}
}

// Extractor pulls values from an [http.Request]
// for domain function parameters.
type Extractor interface {
	RequestValueExtractor
	StringValueExtractor
}

// RequestValueExtractor pulls [url.Values] from an [http.Request]
// in order to provide a [Decoder] with values to populate
// domain request struct with.
type RequestValueExtractor interface {
	ExtractRequestValue(url.Values, *http.Request) error
}

// RequestValueExtractorFunc provides a wrapper for functional
// implementation of a [RequestValueExtractor].
type RequestValueExtractorFunc func(url.Values, *http.Request) error

// RequestValueExtractor satisfies [RequestValueExtractor] interface
// for [RequestValueExtractorFunc].
func (f RequestValueExtractorFunc) ExtractRequestValue(vs url.Values, r *http.Request) error {
	return f(vs, r)
}

// StringValueExtractor pulls out a string value from an [http.Request].
// It is used primarily for custom implentations of
// [htadaptor.UnaryStringFuncAdaptor] and
// [htadaptor.VoidStringFuncAdaptor].
type StringValueExtractor interface {
	ExtractStringValue(*http.Request) (string, error)
}

// StringValueExtractorFunc is a convient function type that
// satisfies [StringValue].
type StringValueExtractorFunc func(*http.Request) (string, error)

// ExtractStringValue satisfies [StringValue] for [StringValueFunc].
func (f StringValueExtractorFunc) ExtractStringValue(r *http.Request) (string, error) {
	return f(r)
}

type Sequence []RequestValueExtractor

func (s Sequence) ExtractRequestValue(vs url.Values, r *http.Request) (err error) {
	for _, extractor := range s {
		if err = extractor.ExtractRequestValue(vs, r); err != nil {
			return err
		}
	}
	return nil
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
		return Sequence(exs)
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
