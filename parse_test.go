package main

import (
	"testing"
)

func TestParseInputFull(t *testing.T) {
	p, err := parseInput([]string{"Buy", "milk", "!high", "@groceries", "#home", "due:tomorrow", "every:week"}, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.text != "Buy milk" {
		t.Errorf("text = %q, want %q", p.text, "Buy milk")
	}
	if p.pri != PriHigh || !p.hasPri {
		t.Errorf("pri = %v (has %v), want high", p.pri, p.hasPri)
	}
	if len(p.tags) != 1 || p.tags[0] != "groceries" {
		t.Errorf("tags = %v, want [groceries]", p.tags)
	}
	if p.project != "home" {
		t.Errorf("project = %q, want home", p.project)
	}
	if p.due != "2026-08-27" {
		t.Errorf("due = %q, want 2026-08-27", p.due)
	}
	if p.every != "week" {
		t.Errorf("every = %q, want week", p.every)
	}
}

func TestParseInputBangPriorities(t *testing.T) {
	cases := map[string]Priority{
		"!": PriLow, "!!": PriMed, "!!!": PriHigh, "!!!!": PriHigh,
		"!h": PriHigh, "!HIGH": PriHigh, "!m": PriMed, "!low": PriLow,
		"!none": PriNone,
	}
	for in, want := range cases {
		p, err := parseInput([]string{"task", in}, testNow)
		if err != nil {
			t.Errorf("parseInput(%q): unexpected error: %v", in, err)
			continue
		}
		if p.pri != want {
			t.Errorf("parseInput(%q): pri = %v, want %v", in, p.pri, want)
		}
	}
}

func TestParseInputQuotedText(t *testing.T) {
	// A single quoted arg still splits into words.
	p, err := parseInput([]string{"Buy oat milk @shop"}, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.text != "Buy oat milk" {
		t.Errorf("text = %q, want %q", p.text, "Buy oat milk")
	}
	if len(p.tags) != 1 || p.tags[0] != "shop" {
		t.Errorf("tags = %v, want [shop]", p.tags)
	}
}

func TestParseInputClears(t *testing.T) {
	p, err := parseInput([]string{"task", "due:none", "every:-", "!none"}, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.hasDue || p.due != "" {
		t.Errorf("due:none should set hasDue with empty due, got has=%v due=%q", p.hasDue, p.due)
	}
	if !p.hasEvery || p.every != "" {
		t.Errorf("every:- should set hasEvery with empty every, got has=%v every=%q", p.hasEvery, p.every)
	}
	if !p.hasPri || p.pri != PriNone {
		t.Errorf("!none should set hasPri with PriNone, got has=%v pri=%v", p.hasPri, p.pri)
	}
}

func TestParseInputErrors(t *testing.T) {
	for _, args := range [][]string{
		{"task", "due:someday"},
		{"task", "every:often"},
		{"task", "!urgent"},
	} {
		if _, err := parseInput(args, testNow); err == nil {
			t.Errorf("parseInput(%v): expected error, got none", args)
		}
	}
}

func TestParseInputBareSymbolsAreText(t *testing.T) {
	p, err := parseInput([]string{"meet", "@", "2pm", "#"}, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.text != "meet @ 2pm #" {
		t.Errorf("text = %q, want %q", p.text, "meet @ 2pm #")
	}
	if p.hasTags || p.hasProj {
		t.Errorf("bare @/# should not set metadata: %+v", p)
	}
}
