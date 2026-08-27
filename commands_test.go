package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseIDs(t *testing.T) {
	s := tempStore(t)
	s.Add(Task{Text: "a"}) // 1
	s.Add(Task{Text: "b"}) // 2
	s.Add(Task{Text: "c"}) // 3
	s.Delete(2)            // gap at 2
	s.Add(Task{Text: "d"}) // reuses 2
	s.Add(Task{Text: "e"}) // 4

	ids, err := parseIDs(s, []string{"1", "3-4"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 3 || ids[2] != 4 {
		t.Errorf("parseIDs = %v, want [1 3 4]", ids)
	}

	// Ranges skip gaps silently; unknown explicit IDs error.
	if _, err := parseIDs(s, []string{"99"}); err == nil {
		t.Error("parseIDs(99): expected error for missing ID")
	}
	if _, err := parseIDs(s, []string{"abc"}); err == nil {
		t.Error("parseIDs(abc): expected error for non-numeric ID")
	}
}

func TestCmdDoneSpawnsRecurrence(t *testing.T) {
	s := tempStore(t)
	s.Add(Task{Text: "water plants", Due: "2026-08-26", Every: "day"})

	var buf bytes.Buffer
	if err := cmdDone(s, &buf, []string{"1"}, testNow); err != nil {
		t.Fatalf("cmdDone: %v", err)
	}

	orig := s.ByID(1)
	if orig == nil || !orig.Done || orig.Completed == nil {
		t.Fatalf("task 1 should be done with a completion time: %+v", orig)
	}
	next := s.ByID(2)
	if next == nil {
		t.Fatal("recurring task should spawn a next occurrence")
	}
	if next.Done || next.Due != "2026-08-27" || next.Every != "day" || next.Text != "water plants" {
		t.Errorf("next occurrence mismatch: %+v", next)
	}
	if !strings.Contains(buf.String(), "#2") {
		t.Errorf("output should mention the spawned task, got: %q", buf.String())
	}
}

func TestCmdAddAndList(t *testing.T) {
	s := tempStore(t)
	var buf bytes.Buffer
	if err := cmdAdd(s, &buf, []string{"Buy", "milk", "!high", "due:tomorrow"}, testNow); err != nil {
		t.Fatalf("cmdAdd: %v", err)
	}
	if !strings.Contains(buf.String(), "#1") {
		t.Errorf("add output should include the new ID, got: %q", buf.String())
	}

	buf.Reset()
	if err := cmdList(s, &buf, nil, testNow); err != nil {
		t.Fatalf("cmdList: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Buy milk") || !strings.Contains(out, "TOMORROW") {
		t.Errorf("list output missing task or group: %q", out)
	}

	// Query filter.
	buf.Reset()
	if err := cmdList(s, &buf, []string{"nomatch"}, testNow); err != nil {
		t.Fatalf("cmdList: %v", err)
	}
	if !strings.Contains(buf.String(), "No tasks match") {
		t.Errorf("expected no-match message, got: %q", buf.String())
	}
}

func TestCmdEdit(t *testing.T) {
	s := tempStore(t)
	s.Add(Task{Text: "old", Priority: PriLow, Tags: []string{"x"}})

	var buf bytes.Buffer
	err := cmdEdit(s, &buf, []string{"1", "new", "text", "!high", "@y", "due:fri"}, testNow)
	if err != nil {
		t.Fatalf("cmdEdit: %v", err)
	}
	got := s.ByID(1)
	if got.Text != "new text" || got.Priority != PriHigh || len(got.Tags) != 1 || got.Tags[0] != "y" || got.Due != "2026-08-28" {
		t.Errorf("edit mismatch: %+v", got)
	}

	// Metadata-only edit keeps the text.
	if err := cmdEdit(s, &buf, []string{"1", "due:none"}, testNow); err != nil {
		t.Fatalf("cmdEdit due:none: %v", err)
	}
	if got := s.ByID(1); got.Text != "new text" || got.Due != "" {
		t.Errorf("due:none should clear due and keep text: %+v", got)
	}
}

func TestCmdClear(t *testing.T) {
	s := tempStore(t)
	s.Add(Task{Text: "a", Done: true})
	s.Add(Task{Text: "b"})

	var buf bytes.Buffer
	// Non-terminal stdin: no confirmation prompt, clears immediately.
	if err := cmdClear(s, &buf, nil, nil); err != nil {
		t.Fatalf("cmdClear: %v", err)
	}
	if len(s.Tasks()) != 1 || s.Tasks()[0].Text != "b" {
		t.Errorf("clear should remove only completed tasks: %+v", s.Tasks())
	}
}

func TestRenderStatsSmoke(t *testing.T) {
	s := tempStore(t)
	c := testNow
	s.Add(Task{Text: "done one", Done: true, Completed: &c})
	s.Add(Task{Text: "pending", Project: "work", Due: "2026-08-25"})

	out := renderStats(s.Tasks(), testNow)
	for _, want := range []string{"Pending", "Completed", "Last 14 days", "Streak", "#work", "overdue"} {
		if !strings.Contains(out, want) {
			t.Errorf("stats output missing %q:\n%s", want, out)
		}
	}
}

func TestStreaks(t *testing.T) {
	byDay := map[string]int{
		"2026-08-26": 1, // today
		"2026-08-25": 2,
		"2026-08-24": 1,
		"2026-08-22": 5, // gap before this
	}
	cur, best := streaks(byDay, today(testNow))
	if cur != 3 {
		t.Errorf("current streak = %d, want 3", cur)
	}
	if best != 3 {
		t.Errorf("best streak = %d, want 3", best)
	}

	// No completions today yet: streak counts through yesterday.
	delete(byDay, "2026-08-26")
	cur, _ = streaks(byDay, today(testNow))
	if cur != 2 {
		t.Errorf("current streak without today = %d, want 2", cur)
	}
}
