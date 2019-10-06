package values

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	timeRegexpString = `^\$\(timenow:(\w*)\)$`
	timeRegexp       = regexp.MustCompile(timeRegexpString)

	formatLayouts = map[string]string{
		"ANSIC":       "Mon Jan _2 15:04:05 2006",
		"UnixDate":    "Mon Jan _2 15:04:05 MST 2006",
		"RubyDate":    "Mon Jan 02 15:04:05 -0700 2006",
		"RFC822":      "02 Jan 06 15:04 MST",
		"RFC822Z":     "02 Jan 06 15:04 -0700",
		"RFC850":      "Monday, 02-Jan-06 15:04:05 MST",
		"RFC1123":     "Mon, 02 Jan 2006 15:04:05 MST",
		"RFC1123Z":    "Mon, 02 Jan 2006 15:04:05 -0700",
		"RFC3339":     "2006-01-02T15:04:05Z07:00",
		"RFC3339Nano": "2006-01-02T15:04:05.999999999Z07:00",
		"Kitchen":     "3:04PM",
		"Stamp":       "Jan _2 15:04:05",
		"StampMilli":  "Jan _2 15:04:05.000",
		"StampMicro":  "Jan _2 15:04:05.000000",
		"StampNano":   "Jan _2 15:04:05.000000000",
	}
)

type nowTimeString struct {
	nowTimeGetter func() time.Time
}

// NowTimeStringParam represents the param of nowtime.
type NowTimeStringParam struct {
	Format string `json:"format"`
}

// Value generates timestamp string for current time based on the input params.
func (t *nowTimeString) Value(param interface{}) string {
	switch v := param.(type) {
	case string:
		return t.Parse(v)
	case NowTimeStringParam:
		return t.value(&v)
	case *NowTimeStringParam:
		return t.value(v)
	default:
		return t.value(&NowTimeStringParam{})
	}
}

func (t *nowTimeString) value(param *NowTimeStringParam) string {
	if len(param.Format) == 0 {
		return strconv.FormatInt(t.nowTimeGetter().Unix(), 10)
	}

	if layout, ok := formatLayouts[param.Format]; ok {
		return t.nowTimeGetter().Format(layout)
	}

	return t.nowTimeGetter().Format(param.Format)
}

// Parse parses now time string value from a string. If the input string is a valid now time string
// ref value: $(timenow:<format>), for example: $(timenow:RFC1123), then the formatted now time, otherwise
// return the input string itself.
// Note, <format> empty, $(timenow:) means generate a timestamp like 1558667413
func (t *nowTimeString) Parse(v string) string {
	trimed := strings.TrimSpace(v)
	results := timeRegexp.FindStringSubmatch(trimed)
	if len(results) < 2 {
		return v
	}

	return t.Value(&NowTimeStringParam{Format: results[1]})
}
