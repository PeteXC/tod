package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// today returns now truncated to local midnight.
func today(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

var weekdays = map[string]time.Weekday{
	"mon": time.Monday, "monday": time.Monday,
	"tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday,
	"wed": time.Wednesday, "weds": time.Wednesday, "wednesday": time.Wednesday,
	"thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday,
	"fri": time.Friday, "friday": time.Friday,
	"sat": time.Saturday, "saturday": time.Saturday,
	"sun": time.Sunday, "sunday": time.Sunday,
}

var relDateRe = regexp.MustCompile(`^\+(\d+)([dwm])$`)
var everyRe = regexp.MustCompile(`^(\d+)([dwm])$`)

// parseDate understands natural due dates: today, tomorrow, yesterday,
// weekday names (fri, friday), relative offsets (+3d, +2w, +1m) and
// ISO dates (2026-09-01). Weekdays resolve to their next occurrence,
// including today.
func parseDate(s string, now time.Time) (time.Time, error) {
	ls := strings.ToLower(strings.TrimSpace(s))
	base := today(now)
	switch ls {
	case "today", "tod":
		return base, nil
	case "tomorrow", "tom":
		return base.AddDate(0, 0, 1), nil
	case "yesterday":
		return base.AddDate(0, 0, -1), nil
	}
	if wd, ok := weekdays[ls]; ok {
		d := base
		for d.Weekday() != wd {
			d = d.AddDate(0, 0, 1)
		}
		return d, nil
	}
	if m := relDateRe.FindStringSubmatch(ls); m != nil {
		n, _ := strconv.Atoi(m[1])
		switch m[2] {
		case "d":
			return base.AddDate(0, 0, n), nil
		case "w":
			return base.AddDate(0, 0, 7*n), nil
		case "m":
			return base.AddDate(0, n, 0), nil
		}
	}
	if t, err := time.ParseInLocation("2006-01-02", ls, now.Location()); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("can't parse date %q (try today, tomorrow, fri, +3d, +2w, or 2026-09-01)", s)
}

// parseEvery validates a recurrence spec and returns its normalized form:
// day, weekday, week, month, a weekday abbreviation (mon..sun), or an
// interval like 3d / 2w / 1m.
func parseEvery(s string) (string, error) {
	ls := strings.ToLower(strings.TrimSpace(s))
	switch ls {
	case "day", "daily":
		return "day", nil
	case "weekday", "weekdays":
		return "weekday", nil
	case "week", "weekly":
		return "week", nil
	case "month", "monthly":
		return "month", nil
	}
	if wd, ok := weekdays[ls]; ok {
		return strings.ToLower(wd.String()[:3]), nil
	}
	if everyRe.MatchString(ls) {
		return ls, nil
	}
	return "", fmt.Errorf("can't parse recurrence %q (try day, weekday, week, month, mon, 3d, or 2w)", s)
}

// everyLabel renders a recurrence spec for display.
func everyLabel(spec string) string {
	switch spec {
	case "day":
		return "daily"
	case "weekday":
		return "weekdays"
	case "week":
		return "weekly"
	case "month":
		return "monthly"
	}
	return "every " + spec
}

// nextOccurrence computes the next due date for a recurring task that was
// just completed. It steps forward from the previous due date until the
// result is strictly in the future: you just did it, so the next occurrence
// is never today. This also collapses long-overdue recurrences into a single
// future occurrence rather than a backlog of catch-up tasks.
func nextOccurrence(due time.Time, spec string, now time.Time) time.Time {
	base := today(now)
	d := due
	if d.IsZero() {
		d = base
	}
	for {
		d = stepOccurrence(d, spec)
		if d.After(base) {
			return d
		}
	}
}

func stepOccurrence(d time.Time, spec string) time.Time {
	switch spec {
	case "day":
		return d.AddDate(0, 0, 1)
	case "week":
		return d.AddDate(0, 0, 7)
	case "month":
		return d.AddDate(0, 1, 0)
	case "weekday":
		n := d.AddDate(0, 0, 1)
		for n.Weekday() == time.Saturday || n.Weekday() == time.Sunday {
			n = n.AddDate(0, 0, 1)
		}
		return n
	}
	if wd, ok := weekdays[spec]; ok {
		n := d.AddDate(0, 0, 1)
		for n.Weekday() != wd {
			n = n.AddDate(0, 0, 1)
		}
		return n
	}
	if m := everyRe.FindStringSubmatch(spec); m != nil {
		n, _ := strconv.Atoi(m[1])
		switch m[2] {
		case "d":
			return d.AddDate(0, 0, n)
		case "w":
			return d.AddDate(0, 0, 7*n)
		case "m":
			return d.AddDate(0, n, 0)
		}
	}
	return d.AddDate(0, 0, 1)
}
