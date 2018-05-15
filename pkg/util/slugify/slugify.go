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

func Slugify(str string, withSuffix bool, maxIDlen int) string {
	s := slugify.Slugify(str)

	if len(s) == 0 {
		s = RandString(9)
	}

	if withSuffix {
		s = AppendRandString(s, randLen)
	}

	if maxIDlen == -1 {
		return s
	}

	if len(s) < 2 {
		s = AppendRandString(s, 9)
	}

	if maxIDlen > randLen+1 && len(s) > maxIDlen {
		cutoff := maxIDlen - randLen - 1
		return AppendRandString(s[:cutoff], randLen)
	}

	return s
}

func AppendRandString(s string, n int) string {
	if strings.HasSuffix(s, "-") || len(s) == 0 {
		return s + RandString(n)
	}
	return s + "-" + RandString(n)
}
