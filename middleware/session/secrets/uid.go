package secrets

import (
	"math/rand"
	"time"
)

const (
	letterBytes   = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits

	UniqueIdentifierCharacterSet = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	VisuallyUnambiguousLowerCaseCharacterSet = `abcdefhijkmnoprstwxy34`
	VisuallyUnambiguousCharacterSet          = VisuallyUnambiguousLowerCaseCharacterSet + `ABCDEFHIJKMNOPRSTWXY`
)

var src = rand.NewSource(time.Now().UnixNano())

// FastRandom is inspired by Ketan Parmar's work:
//
// - https://github.com/kpbird/golang_random_string/blob/master/main.go
// - https://kpbird.medium.com/golang-generate-fixed-size-random-string-dd6dbd5e63c0
func NewID(n int) []byte {
	b := make([]byte, n)
	l := len(letterBytes)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < l {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}
