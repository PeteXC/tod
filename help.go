package main

import (
	"strings"
)

func renderHelp() string {
	var b strings.Builder
	h := stHead.Render
	dim := stDim.Render
	ac := stAccent.Render

	b.WriteString("\n")
	b.WriteString("  " + h("tod") + dim(" — a fast, beautiful to-do list for your terminal") + "\n")

	b.WriteString("\n  " + h("Usage") + "\n")
	b.WriteString("    " + ac("tod") + dim("                        interactive mode (plain list when piped)") + "\n")
	b.WriteString("    " + ac("tod add") + dim(" <text> [meta]     add a task") + "\n")
	b.WriteString("    " + ac("tod ls") + dim(" [filters]         list tasks") + "\n")
	b.WriteString("    " + ac("tod done") + dim(" <id>...           complete tasks (ranges ok: 1-3)") + "\n")
	b.WriteString("    " + ac("tod undone") + dim(" <id>...           reopen tasks") + "\n")
	b.WriteString("    " + ac("tod rm") + dim(" <id>...           delete tasks") + "\n")
	b.WriteString("    " + ac("tod edit") + dim(" <id> <changes>    edit text and metadata") + "\n")
	b.WriteString("    " + ac("tod pri") + dim(" <id> <level>      set priority: high, medium, low, none") + "\n")
	b.WriteString("    " + ac("tod due") + dim(" <id> <when>       set due date (none clears)") + "\n")
	b.WriteString("    " + ac("tod search") + dim(" <query>           find tasks (same as: tod ls <query>)") + "\n")
	b.WriteString("    " + ac("tod stats") + dim("                   your productivity dashboard") + "\n")
	b.WriteString("    " + ac("tod undo") + dim(" / ") + ac("redo") + dim("          every change is undoable") + "\n")
	b.WriteString("    " + ac("tod clear") + dim(" [--force]         remove completed tasks") + "\n")
	b.WriteString("    " + ac("tod export") + dim("                  print all tasks as JSON") + "\n")
	b.WriteString("    " + ac("tod completion") + dim(" <bash|zsh|fish>   shell completion script") + "\n")
	b.WriteString("    " + ac("tod path") + dim(" / ") + ac("version") + dim(" / ") + ac("help") + "\n")

	b.WriteString("\n  " + h("Metadata") + dim(" — inline, in any order, when adding or editing") + "\n")
	b.WriteString("    " + ac("!high !med !low") + dim("     priority (or ") + ac("!!! !! !") + dim(")") + "\n")
	b.WriteString("    " + ac("@tag") + dim("                  tag (repeatable)") + "\n")
	b.WriteString("    " + ac("#project") + dim("              project") + "\n")
	b.WriteString("    " + ac("due:<when>") + dim("            today, tomorrow, fri, +3d, +2w, 2026-09-01") + "\n")
	b.WriteString("    " + ac("every:<span>") + dim("          day, weekday, week, month, mon, 3d, 2w") + "\n")

	b.WriteString("\n  " + h("Filters") + dim(" — for ls and search") + "\n")
	b.WriteString("    " + ac("--all -a") + dim("              include completed (last 10)") + "\n")
	b.WriteString("    " + ac("--done -d") + dim("             completed only") + "\n")
	b.WriteString("    " + ac("@tag  #project") + dim("        filter by tag or project") + "\n")
	b.WriteString("    " + ac("--pri high") + dim("            filter by priority") + "\n")
	b.WriteString("    " + ac("--plain") + dim("               ASCII output for scripts") + "\n")

	b.WriteString("\n  " + h("Examples") + "\n")
	b.WriteString(dim(`    tod add "Water the plants" every:day`) + "\n")
	b.WriteString(dim(`    tod add "Submit report" #work !high due:fri`) + "\n")
	b.WriteString(dim(`    tod add "Buy milk" @groceries due:tomorrow`) + "\n")
	b.WriteString(dim(`    tod done 1-3`) + "\n")
	b.WriteString(dim(`    tod ls @home --all`) + "\n")
	b.WriteString(dim(`    tod edit 2 "Buy oat milk instead" due:+1w`) + "\n")

	b.WriteString("\n  " + h("Interactive keys") + "\n")
	b.WriteString(dim("    space done · a add · e edit · d delete · u undo · / filter") + "\n")
	b.WriteString(dim("    tab show all · 1/2/3 priority · t due today · T tomorrow · q quit") + "\n")

	b.WriteString("\n  " + dim("Data lives in ") + ac("~/.tod") + dim(" (override with $TOD_HOME).") + "\n\n")
	return b.String()
}
