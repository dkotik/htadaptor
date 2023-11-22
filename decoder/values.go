package decoder

import "net/url"

func mergeURLValues(b, a url.Values) {
	for key, valueSet := range a {
		b[key] = valueSet
	}
}
