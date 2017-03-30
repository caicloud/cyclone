package when

import (
	"bytes"
	"fmt"
	"time"
)

var longDayNames = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

var shortDayNames = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

var shortMonthNames = []string{
	"---",
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

var longMonthNames = []string{
	"---",
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

func weekNumber(t *time.Time, char int) int {
	weekday := int(t.Weekday())

	if char == 'W' {
		// Monday as the first day of the week
		if weekday == 0 {
			weekday = 6
		} else {
			weekday--
		}
	}

	return (t.YearDay() + 6 - weekday) / 7
}

// Strftime formats time.Date according to the directives in the given format string. The directives begins with a percent (%) character.
func Strftime(t *time.Time, f string) string {
	// var result []string
	var buf = &bytes.Buffer{}
	format := []rune(f)
	add := func(str string) {
		// result = append(result, str)
		buf.WriteString(str)
	}
	for i := 0; i < len(format); i++ {
		switch format[i] {
		case '%':
			if i < len(format)-1 {
				switch format[i+1] {
				case 'a':
					add(shortDayNames[t.Weekday()])
				case 'A':
					add(longDayNames[t.Weekday()])
				case 'w':
					add(fmt.Sprintf("%d", t.Weekday()))
				case 'd':
					add(fmt.Sprintf("%02d", t.Day()))
				case 'b':
					add(shortMonthNames[t.Month()])
				case 'B':
					add(longMonthNames[t.Month()])
				case 'm':
					add(fmt.Sprintf("%02d", t.Month()))
				case 'y':
					add(fmt.Sprintf("%02d", t.Year()%100))
				case 'Y':
					add(fmt.Sprintf("%02d", t.Year()))
				case 'H':
					add(fmt.Sprintf("%02d", t.Hour()))
				case 'I':
					if t.Hour() == 0 {
						add(fmt.Sprintf("%02d", 12))
					} else if t.Hour() > 12 {
						add(fmt.Sprintf("%02d", t.Hour()-12))
					} else {
						add(fmt.Sprintf("%02d", t.Hour()))
					}
				case 'p':
					if t.Hour() < 12 {
						add("AM")
					} else {
						add("PM")
					}
				case 'M':
					add(fmt.Sprintf("%02d", t.Minute()))
				case 'S':
					add(fmt.Sprintf("%02d", t.Second()))
				case 'f':
					add(fmt.Sprintf("%06d", t.Nanosecond()/1000))
				case 'z':
					add(t.Format("-0700"))
				case 'Z':
					add(t.Format("MST"))
				case 'j':
					add(fmt.Sprintf("%03d", t.YearDay()))
				case 'U':
					add(fmt.Sprintf("%02d", weekNumber(t, 'U')))
				case 'W':
					add(fmt.Sprintf("%02d", weekNumber(t, 'W')))
				case 'c':
					add(t.Format("Mon Jan 2 15:04:05 2006"))
				case 'x':
					add(fmt.Sprintf("%02d/%02d/%02d", t.Month(), t.Day(), t.Year()%100))
				case 'X':
					add(fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()))
				case '%':
					add("%")
				}
				i++
			}
		default:
			add(string(format[i]))
		}
	}

	// return strings.Join(result, "")
	return buf.String()
}
