package collector

import "time"

func TimeBetween(t, start, end time.Time) bool {
	if start.After(end) {
		start, end = end, start
	}
	return (t.Equal(start) || t.After(start)) &&
		(t.Equal(end) || t.Before(end))
}
