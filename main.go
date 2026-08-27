package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

const version = "1.0.0"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	now := time.Now()

	// Leading global flags.
	for len(args) > 0 && strings.HasPrefix(args[0], "-") && args[0] != "-" {
		switch args[0] {
		case "--plain":
			setPlain()
		case "--no-color":
			setNoColor()
		case "-h", "--help":
			fmt.Print(renderHelp())
			return 0
		case "-v", "--version":
			fmt.Printf("tod %s\n", version)
			return 0
		default:
			fmt.Fprintf(os.Stderr, "%s %s\n\nRun `tod help` for usage.\n", stError.Render("unknown flag:"), args[0])
			return 2
		}
		args = args[1:]
	}

	store, err := OpenStore(dataDir())
	if err != nil {
		fmt.Fprintln(os.Stderr, stError.Render("error: ")+err.Error())
		return 1
	}

	// Bare `tod`: interactive UI on a terminal, plain list when piped.
	if len(args) == 0 {
		if isInteractive() {
			return runTUI(store)
		}
		fmt.Print(renderList(store.Tasks(), listOpts{now: now, width: termWidth()}))
		return 0
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "add", "a":
		err = cmdAdd(store, os.Stdout, rest, now)
	case "ls", "list", "l", "search", "find", "f":
		err = cmdList(store, os.Stdout, rest, now)
	case "done", "do", "check":
		err = cmdDone(store, os.Stdout, rest, now)
	case "undone", "reopen":
		err = cmdUndone(store, os.Stdout, rest, now)
	case "rm", "del", "delete":
		err = cmdRm(store, os.Stdout, rest, now)
	case "edit", "e":
		err = cmdEdit(store, os.Stdout, rest, now)
	case "pri", "priority":
		err = cmdPri(store, os.Stdout, rest, now)
	case "due":
		err = cmdDue(store, os.Stdout, rest, now)
	case "stats":
		err = cmdStats(store, os.Stdout, now)
	case "undo", "u":
		err = cmdUndo(store, os.Stdout)
	case "redo":
		err = cmdRedo(store, os.Stdout)
	case "clear":
		err = cmdClear(store, os.Stdout, rest, os.Stdin)
	case "export":
		err = cmdExport(store, os.Stdout)
	case "path":
		cmdPath(store, os.Stdout)
	case "completion":
		err = cmdCompletion(os.Stdout, rest)
	case "help", "-h", "--help":
		fmt.Print(renderHelp())
	case "version", "-v", "--version":
		fmt.Printf("tod %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "%s %s\n\nRun `tod help` for usage.\n", stError.Render("unknown command:"), cmd)
		return 2
	}
	if err != nil {
		if ue, ok := err.(usageErr); ok {
			fmt.Fprintln(os.Stderr, stError.Render("usage: ")+ue.Error())
			return 2
		}
		fmt.Fprintln(os.Stderr, stError.Render("error: ")+err.Error())
		return 1
	}
	return 0
}

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func termWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 80
}

func runTUI(s *Store) int {
	m := newTUIModel(s)
	m.refresh()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error running interactive mode: "+err.Error())
		return 1
	}
	return 0
}
