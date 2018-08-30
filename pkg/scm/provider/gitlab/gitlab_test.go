package gitlab

import (
	"testing"
)

func TestGetTopLanguage(t *testing.T) {

	testCases := map[string]struct {
		languages map[string]float32
		top       string
	}{
		"empty": {
			make(map[string]float32),
			"",
		},
		"normal": {
			map[string]float32{"Java": 59.9, "Go": 40.1},
			"Java",
		},
	}

	for d, tc := range testCases {
		language := getTopLanguage(tc.languages)
		if language != tc.top {
			t.Errorf("%s failed as error : Expect result %s equals to %s", d, tc.top, language)
		}
	}

}
