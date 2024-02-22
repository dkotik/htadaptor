package session

import "fmt"

type Error uint8

const (
	ErrNoSessionInContext Error = iota
	ErrLargeCookie
)

func (e Error) Error() string {
	switch e {
	case ErrNoSessionInContext:
		return "no session in context"
	case ErrLargeCookie:
		return fmt.Sprintf("cookie must be less than %d bytes long", MaximumCookieSize)
	default:
		return "unknown session error"
	}
}
