package github

import "testing"

func TestGetTopLanguage(t *testing.T) {

	testCases := map[string]struct {
		languages map[string]int
		top       string
	}{
		"empty": {
			make(map[string]int),
			"",
		},
		"normal": {
			map[string]int{"Java": 59, "Go": 41},
			"Java",
		},
		"same value": {
			map[string]int{"Java": 50.0, "Go": 50.0},
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
