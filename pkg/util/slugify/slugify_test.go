/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package slugify

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	testCases := []struct {
		str        string
		withSuffix bool
		maxIDlen   int
		expect     string
	}{
		{
			str:        "qwert-qwert-qwert-qwert",
			withSuffix: true,
			maxIDlen:   12,
		},
		{
			str:        "qwert-qwert",
			withSuffix: true,
			maxIDlen:   12,
		},
		{
			str:        "qwert-qwert",
			withSuffix: false,
			maxIDlen:   12,
		},
		{
			str:        "qwert-qwert-qwert-qwert",
			withSuffix: false,
			maxIDlen:   12,
		},
		{
			str:        "%%%%%%%%%%%%",
			withSuffix: false,
			maxIDlen:   12,
		},
		{
			str:        "我很长很长很长。。。",
			withSuffix: false,
			maxIDlen:   12,
		},
		{
			str:        "%aa%",
			withSuffix: true,
			maxIDlen:   12,
		},
		{
			str:        "@a@",
			withSuffix: false,
			maxIDlen:   12,
		},
		{
			str:        "----",
			withSuffix: false,
			maxIDlen:   12,
		},
		{
			str:        "abcdefghijk",
			withSuffix: true,
			maxIDlen:   12,
		},
		{
			str:        "abcdefghijklmnopqrst",
			withSuffix: true,
			maxIDlen:   12,
		},
		{
			str:        "a",
			withSuffix: true,
			maxIDlen:   12,
		},
	}

	for _, testCase := range testCases {
		result := Slugify(testCase.str, testCase.withSuffix, testCase.maxIDlen)
		t.Log(result)
		if len(result) > 12 || len(result) < 2 {
			t.Errorf("result should in 2 ~ 12")
		}
	}
}
