package main

import (
	"testing"
	"time"
)

// now is Wednesday, 2026-08-26.
var testNow = time.Date(2026, 8, 26, 15, 30, 0, 0, time.Local)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func TestParseDate(t *testing.T) {
	cases := []struct {
		in   string
		want time.Time
	}{
		{"today", day(2026, 8, 26)},
		{"TODAY", day(2026, 8, 26)},
		{"tomorrow", day(2026, 8, 27)},
		{"tom", day(2026, 8, 27)},
		{"yesterday", day(2026, 8, 25)},
		{"wed", day(2026, 8, 26)}, // same weekday resolves to today
		{"wednesday", day(2026, 8, 26)},
		{"thu", day(2026, 8, 27)},
		{"fri", day(2026, 8, 28)},
		{"mon", day(2026, 8, 31)}, // next Monday
		{"sunday", day(2026, 8, 30)},
		{"+0d", day(2026, 8, 26)},
		{"+3d", day(2026, 8, 29)},
		{"+2w", day(2026, 9, 9)},
		{"+1m", day(2026, 9, 26)},
		{"2026-12-25", day(2026, 12, 25)},
	}
	for _, c := range cases {
		got, err := parseDate(c.in, testNow)
		if err != nil {
			t.Errorf("parseDate(%q): unexpected error: %v", c.in, err)
			continue
		}
		if !got.Equal(c.want) {
			t.Errorf("parseDate(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseDateErrors(t *testing.T) {
	for _, in := range []string{"", "nope", "32-13-99", "+5y", "2026/01/01"} {
		if _, err := parseDate(in, testNow); err == nil {
			t.Errorf("parseDate(%q): expected error, got none", in)
		}
	}
}

func TestParseEvery(t *testing.T) {
	cases := map[string]string{
		"day": "day", "daily": "day",
		"weekday": "weekday", "weekdays": "weekday",
		"week": "week", "weekly": "week",
		"month": "month", "monthly": "month",
		"mon": "mon", "Monday": "mon", "FRIDAY": "fri",
		"3d": "3d", "2w": "2w", "1m": "1m",
	}
	for in, want := range cases {
		got, err := parseEvery(in)
		if err != nil {
			t.Errorf("parseEvery(%q): unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parseEvery(%q) = %q, want %q", in, got, want)
		}
	}
	for _, in := range []string{"", "sometimes", "0x", "every"} {
		if _, err := parseEvery(in); err == nil {
			t.Errorf("parseEvery(%q): expected error, got none", in)
		}
	}
}

func TestNextOccurrence(t *testing.T) {
	cases := []struct {
		name string
		due  time.Time
		spec string
		want time.Time
	}{
		{"daily from today", day(2026, 8, 26), "day", day(2026, 8, 27)},
		{"weekly from today", day(2026, 8, 26), "week", day(2026, 9, 2)},
		{"monthly", day(2026, 8, 26), "month", day(2026, 9, 26)},
		// Overdue: steps forward until strictly future (never today —
		// you just completed it).
		{"weekly overdue catch-up", day(2026, 8, 19), "week", day(2026, 9, 2)},
		{"daily overdue catch-up", day(2026, 8, 20), "day", day(2026, 8, 27)},
		// Weekday recurrence skips the weekend: Fri -> Mon.
		{"weekday from friday", day(2026, 8, 28), "weekday", day(2026, 8, 31)},
		{"weekday from wednesday", day(2026, 8, 26), "weekday", day(2026, 8, 27)},
		// Named weekday: next Monday strictly after due.
		{"every monday", day(2026, 8, 26), "mon", day(2026, 8, 31)},
		{"every wednesday from wed", day(2026, 8, 26), "wed", day(2026, 9, 2)},
		{"every 3d", day(2026, 8, 26), "3d", day(2026, 8, 29)},
	}
	for _, c := range cases {
		got := nextOccurrence(c.due, c.spec, testNow)
		if !got.Equal(c.want) {
			t.Errorf("%s: nextOccurrence(%v, %q) = %v, want %v", c.name, c.due, c.spec, got, c.want)
		}
	}
}

func TestNextOccurrenceZeroDue(t *testing.T) {
	// A recurring task with no due date starts counting from today.
	got := nextOccurrence(time.Time{}, "day", testNow)
	if !got.Equal(day(2026, 8, 27)) {
		t.Errorf("nextOccurrence(zero, day) = %v, want 2026-08-27", got)
	}
}
