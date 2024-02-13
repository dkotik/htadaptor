/*
Package middleware includes some of the most frequently used HTTP middleware handlers for the the following reasons:

1. reduce the number of project utility dependencies
2. keep logging more consistent
3. keep error handling more consistent
*/
package middleware

import "net/http"

// TODO: Middleware func(http.Handler) (http.Handler, error)
type Middleware func(http.Handler) http.Handler

// Must panics if middleware creation returned an error.
func Must(h http.Handler, err error) http.Handler {
	if err != nil {
		panic(err)
	}
	return h
}
