package matcher

import (
	"regexp"
	"strings"
)

var splitter = regexp.MustCompile(`[:=] ?`)
var pwdPattern = regexp.MustCompile(`(?i)(?:^|\W)(?:pwd|password)(?:=|: ?)([^\s:=]+)`)
var ipPattern = regexp.MustCompile(`\d+\.\d+\.\d+.\d+`)

func MaskPwd(msg string) string {
	result := msg
	for _, m := range pwdPattern.FindAllStringSubmatch(msg, -1) {
		p := splitter.Split(m[0], -1)
		t := strings.Replace(m[0], p[len(p)-1], "******", -1)
		result = strings.Replace(result, m[0], t, -1)
	}

	return result
}

func IsIP(value string) bool {
	return ipPattern.MatchString(value)
}
