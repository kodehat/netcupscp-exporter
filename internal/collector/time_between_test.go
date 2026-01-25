package collector

import (
	"testing"
	"time"
)

func TestTimeBetween(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.UTC().Location())

	tests := []struct {
		name          string
		t, start, end time.Time
		want          bool
	}{
		{
			"inside",
			today.Add(10 * time.Hour),
			today.Add(9 * time.Hour),
			today.Add(13 * time.Hour),
			true,
		},
		{
			"equal start",
			today.Add(9 * time.Hour),
			today.Add(9 * time.Hour),
			today.Add(13 * time.Hour),
			true,
		},
		{
			"equal end",
			today.Add(13 * time.Hour),
			today.Add(9 * time.Hour),
			today.Add(13 * time.Hour),
			true,
		},
		{
			"before start",
			today.Add(8 * time.Hour),
			today.Add(9 * time.Hour),
			today.Add(13 * time.Hour),
			false,
		},
		{
			"after end",
			today.Add(14 * time.Hour),
			today.Add(9 * time.Hour),
			today.Add(13 * time.Hour),
			false,
		},
		{
			"swapped bounds",
			today.Add(10 * time.Hour),
			today.Add(13 * time.Hour),
			today.Add(9 * time.Hour),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeBetween(tt.t, tt.start, tt.end); got != tt.want {
				t.Errorf("TimeBetween() = %v; want %v", got, tt.want)
			}
		})
	}
}
