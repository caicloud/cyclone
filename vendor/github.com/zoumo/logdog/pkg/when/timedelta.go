package when

import (
	"time"
)

func abs(v time.Duration) time.Duration {
	if v < 0 {
		v *= -1
	}
	return v
}

// Timedelta represents a duration between two dates.
// All fields are optional and default to 0. You can initialize any type of timedelta by specifying field values which you want to use.
type Timedelta struct {
	Days, Seconds, Microseconds, Milliseconds, Minutes, Hours, Weeks time.Duration
}

// Add returns the Timedelta t+t2.
func (t *Timedelta) Add(t2 *Timedelta) Timedelta {
	return Timedelta{
		Days:         t.Days + t2.Days,
		Seconds:      t.Seconds + t2.Seconds,
		Microseconds: t.Microseconds + t2.Microseconds,
		Milliseconds: t.Milliseconds + t2.Milliseconds,
		Minutes:      t.Minutes + t2.Minutes,
		Hours:        t.Hours + t2.Hours,
		Weeks:        t.Weeks + t2.Weeks,
	}
}

// Subtract returns the Timedelta t-t2.
func (t *Timedelta) Subtract(t2 *Timedelta) Timedelta {
	return Timedelta{
		Days:         t.Days - t2.Days,
		Seconds:      t.Seconds - t2.Seconds,
		Microseconds: t.Microseconds - t2.Microseconds,
		Milliseconds: t.Milliseconds - t2.Milliseconds,
		Minutes:      t.Minutes - t2.Minutes,
		Hours:        t.Hours - t2.Hours,
		Weeks:        t.Weeks - t2.Weeks,
	}
}

// Abs returns the absolute value of t
func (t *Timedelta) Abs() Timedelta {
	return Timedelta{
		Days:         abs(t.Days),
		Seconds:      abs(t.Seconds),
		Microseconds: abs(t.Microseconds),
		Milliseconds: abs(t.Milliseconds),
		Minutes:      abs(t.Minutes),
		Hours:        abs(t.Hours),
		Weeks:        abs(t.Weeks),
	}
}

// Duration returns time.Duration. time.Duration can be added to time.Date.
func (t *Timedelta) Duration() time.Duration {
	return t.Days*24*time.Hour +
		t.Seconds*time.Second +
		t.Microseconds*time.Microsecond +
		t.Milliseconds*time.Millisecond +
		t.Minutes*time.Minute +
		t.Hours*time.Hour +
		t.Weeks*7*24*time.Hour
}

// String returns a string representing the Timedelta's duration in the form "72h3m0.5s".
func (t *Timedelta) String() string {
	return t.Duration().String()
}
