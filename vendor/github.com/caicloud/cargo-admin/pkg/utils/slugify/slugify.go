package slugify

import (
	"math/rand"
	"time"

	slugify "github.com/mozillazg/go-slugify"
)

const randLen = 7

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func Slugify(str string, withSuffix bool) string {
	s := slugify.Slugify(str)
	if withSuffix {
		return s + "-" + RandString(randLen)
	}
	return s
}
