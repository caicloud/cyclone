/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package slugify

import (
	"math/rand"
	"strings"
	"time"

	"github.com/mozillazg/go-slugify"
)

const randLen = 5

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

var src = rand.NewSource(time.Now().UnixNano())

// randString return a string with n characters
func randString(n int) string {
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

// Slugify converts input 'str' to a string only containing alphameric and "-".
// if withSuffix is true, will append a random string
// maxIDlen controls the max length of the string
func Slugify(str string, withSuffix bool, maxIDlen int) string {
	s := slugify.Slugify(str)

	if len(s) == 0 {
		s = randString(9)
	}

	if withSuffix {
		s = appendRandString(s, randLen)
	}

	if maxIDlen == -1 {
		return s
	}

	if len(s) < 2 {
		s = appendRandString(s, 9)
	}

	if maxIDlen > randLen+1 && len(s) > maxIDlen {
		cutoff := maxIDlen - randLen - 1
		return appendRandString(s[:cutoff], randLen)
	}

	return s
}

func appendRandString(s string, n int) string {
	if strings.HasSuffix(s, "-") || len(s) == 0 {
		return s + randString(n)
	}
	return s + "-" + randString(n)
}
