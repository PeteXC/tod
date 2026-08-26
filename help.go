package main

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

func renderHelp() string {
	var b strings.Builder
	h := stHead.Render
	dim := stDim.Render
	ac := stAccent.Render

	section := func(title string, rows [][2]string) {
		b.WriteString("\n  " + h(title) + "\n")
		w := 0
		for _, r := range rows {
			if rw := runewidth.StringWidth(r[0]); rw > w {
				w = rw
			}
		}
		for _, r := range rows {
			if r[1] == "" {
				b.WriteString("    " + dim(r[0]) + "\n")
				continue
			}
			b.WriteString("    " + ac(r[0]) + strings.Repeat(" ", w-runewidth.StringWidth(r[0])) + dim("  "+r[1]) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString("  " + h("tod") + dim(" — a fast, beautiful to-do list for your terminal") + "\n")

	section("Usage", [][2]string{
		{"tod", "interactive mode (plain list when piped)"},
		{"tod add <text> [meta]", "add a task"},
		{"tod ls [filters]", "list tasks"},
		{"tod done <id>...", "complete tasks (ranges ok: 1-3)"},
		{"tod undone <id>...", "reopen tasks"},
		{"tod rm <id>...", "delete tasks"},
		{"tod edit <id> <changes>", "edit text and metadata"},
		{"tod pri <id> <level>", "set priority: high, medium, low, none"},
		{"tod due <id> <when>", "set due date (none clears)"},
		{"tod search <query>", "find tasks (same as: tod ls <query>)"},
		{"tod stats", "your productivity dashboard"},
		{"tod undo / redo", "every change is undoable"},
		{"tod clear [--force]", "remove completed tasks"},
		{"tod export", "print all tasks as JSON"},
		{"tod completion <shell>", "bash, zsh, or fish"},
		{"tod path / version / help", "data locations, version, this page"},
	})

	section("Metadata — inline, in any order, when adding or editing", [][2]string{
		{"!high !med !low", "priority (or !!! !! !)"},
		{"@tag", "tag (repeatable)"},
		{"#project", "project"},
		{"due:<when>", "today, tomorrow, fri, +3d, +2w, 2026-09-01"},
		{"every:<span>", "day, weekday, week, month, mon, 3d, 2w"},
	})

	section("Filters — for ls and search", [][2]string{
		{"--all -a", "include completed (last 10)"},
		{"--done -d", "completed only"},
		{"@tag  #project", "filter by tag or project"},
		{"--pri high", "filter by priority"},
		{"--plain", "ASCII output for scripts"},
	})

	section("Examples", [][2]string{
		{`tod add "Water plants" every:day`, ""},
		{`tod add "Submit report" '#work' '!!!' due:fri`, ""},
		{`tod add "Buy milk" @groceries due:tomorrow`, ""},
		{"tod done 1-3", ""},
		{"tod ls @home --all", ""},
		{`tod edit 2 "Buy oat milk" due:+1w`, ""},
	})

	section("Interactive keys", [][2]string{
		{"space a e d u /", "done · add · edit · delete · undo · filter"},
		{"tab 1/2/3 t T q", "show all · priority · due today · tomorrow · quit"},
	})

	b.WriteString("\n  " + dim("Data lives in ") + ac("~/.tod") + dim(" (override with $TOD_HOME).") + "\n\n")
	return b.String()
}
