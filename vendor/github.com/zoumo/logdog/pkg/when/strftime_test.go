package when

import (
	"testing"
	"time"
)

func TestStrftime(t *testing.T) {
	date := time.Date(2005, 2, 3, 4, 5, 6, 7000, time.UTC)
	AssertEqual(t, Strftime(&date,
		"%a %A %w %d %b %B %m %y %Y %H %I %p %M %S %f %z %Z %j %U %W %c %x %X %%"),
		"Thu Thursday 4 03 Feb February 02 05 2005 04 04 AM 05 06 000007 +0000 UTC 034 05 05 Thu Feb 3 04:05:06 2005 02/03/05 04:05:06 %")

	date = time.Date(2015, 7, 2, 15, 24, 30, 35, time.UTC)
	AssertEqual(t, Strftime(&date, "%U %W"), "26 26")

	date = time.Date(1962, 3, 23, 15, 24, 30, 35, time.UTC)
	AssertEqual(t, Strftime(&date, "%U %W"), "11 12")

	date = time.Date(1989, 12, 31, 15, 24, 30, 35000, time.UTC)
	AssertEqual(t, Strftime(&date, "%U %W"), "53 52")

	AssertEqual(t, Strftime(&date,
		"%a %A %w %d %b %B %m %y %Y %H %I %p %M %S %f %z %Z %j %U %W %c %x %X %%"),
		"Sun Sunday 0 31 Dec December 12 89 1989 15 03 PM 24 30 000035 +0000 UTC 365 53 52 Sun Dec 31 15:24:30 1989 12/31/89 15:24:30 %")

	date = time.Date(1989, 12, 31, 0, 24, 30, 35000, time.UTC)
	AssertEqual(t, Strftime(&date, "%I"), "12")

	AssertEqual(t, Strftime(&date, "%a %A %w %d %b %B %"), "Sun Sunday 0 31 Dec December ")

	AssertEqual(t, Strftime(&date, "작성일 : %a %A %w %d %b %B %"), "작성일 : Sun Sunday 0 31 Dec December ")
}
