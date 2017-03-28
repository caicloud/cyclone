package when

import (
	"reflect"
	"testing"
	"time"
)

func AssertEqual(t *testing.T, x, y interface{}) {
	if !reflect.DeepEqual(x, y) {
		t.Error("Expected ", y, ", got ", x)
	}
}

func TestTimedelta(t *testing.T) {
	base := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)
	result := base.Add((&Timedelta{Days: 1, Seconds: 66355, Weeks: 1722}).Duration())
	AssertEqual(t, result.String(), "2013-01-07 18:25:55 +0000 UTC")

	result = result.Add((&Timedelta{Microseconds: 3, Milliseconds: 10, Minutes: 1}).Duration())
	AssertEqual(t, result.String(), "2013-01-07 18:26:55.010003 +0000 UTC")

	td := Timedelta{Days: 10, Minutes: 17, Seconds: 56}
	td2 := Timedelta{Days: 15, Minutes: 13, Seconds: 42}
	td = td.Add(&td2)

	base = time.Date(2015, 2, 3, 0, 0, 0, 0, time.UTC)
	result = base.Add(td.Duration())
	AssertEqual(t, result.String(), "2015-02-28 00:31:38 +0000 UTC")

	td = td.Subtract(&td2)

	result = base.Add(td.Duration())
	AssertEqual(t, result.String(), "2015-02-13 00:17:56 +0000 UTC")

	AssertEqual(t, td.String(), "240h17m56s")

	td = Timedelta{Days: -1, Seconds: -1, Microseconds: -1, Milliseconds: -1, Minutes: -1, Hours: -1, Weeks: -1}
	td2 = td
	td = td.Abs()
	td = td.Add(&td2)
	AssertEqual(t, td.String(), "0s")
}
