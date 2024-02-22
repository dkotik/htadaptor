package extract

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/dkotik/htadaptor/middleware/session"
)

var (
	_ RequestValueExtractor = (singleSessionValue)("")
	_ StringValueExtractor  = (singleSessionValue)("")
	_ RequestValueExtractor = (multiSessionValue)(nil)
	_ StringValueExtractor  = (multiSessionValue)(nil)
)

// NewSessionValueExtractor is a [Extractor] extractor that pulls
// out [session.Session] values by key name from an [http.Request]
// context.
func NewSessionValueExtractor(keys ...string) (Extractor, error) {
	total := len(keys)
	if total == 0 {
		return nil, errors.New("session value extractor requires at least one parameter name")
	}
	if err := uniqueNonEmptyValueNames(keys); err != nil {
		return nil, err
	}
	if total == 1 {
		return singleSessionValue(keys[0]), nil
	}
	return multiSessionValue(keys), nil
}

type singleSessionValue string

func (e singleSessionValue) ExtractRequestValue(vs url.Values, r *http.Request) error {
	desired := string(e)
	if strValue, ok := session.Value(r.Context(), desired).(string); ok && len(strValue) > 0 {
		vs[desired] = []string{strValue}
	}
	return nil
}

func (e singleSessionValue) ExtractStringValue(r *http.Request) (string, error) {
	desired := string(e)
	if strValue, ok := session.Value(r.Context(), desired).(string); ok && len(strValue) > 0 {
		return strValue, nil
	}
	return "", NoValueError{desired}
}

type multiSessionValue []string

func (e multiSessionValue) ExtractRequestValue(vs url.Values, r *http.Request) error {
	return session.Read(r.Context(), func(s session.Session) error {
		for _, desired := range e {
			if strValue, ok := s.Get(desired).(string); ok && len(strValue) > 0 {
				vs[desired] = []string{strValue}
			}
		}
		return nil
	})
}

func (e multiSessionValue) ExtractStringValue(r *http.Request) (result string, err error) {
	err = session.Read(r.Context(), func(s session.Session) error {
		for _, desired := range e {
			if strValue, ok := s.Get(desired).(string); ok && len(strValue) > 0 {
				result = strValue
				return nil
			}
		}
		return NoValueError(e)
	})
	return result, err
}
