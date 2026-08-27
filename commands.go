package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// usageErr marks errors caused by wrong command usage (exit code 2).
type usageErr string

func (e usageErr) Error() string { return string(e) }

func usageError(format string, args ...interface{}) error {
	return usageErr(fmt.Sprintf(format, args...))
}

// parseIDs resolves args like "1", "3", "1-4" to existing task IDs.
// Gaps inside ranges are skipped; an explicitly named ID must exist.
func parseIDs(s *Store, args []string) ([]int, error) {
	var ids []int
	seen := map[int]bool{}
	add := func(id int) {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for _, a := range args {
		if lo, hi, ok := parseRange(a); ok {
			for i := lo; i <= hi; i++ {
				if s.ByID(i) != nil {
					add(i)
				}
			}
			continue
		}
		n, err := strconv.Atoi(a)
		if err != nil || n < 1 {
			return nil, usageError("%q is not a task ID (use the numbers from `tod ls`)", a)
		}
		if s.ByID(n) == nil {
			return nil, fmt.Errorf("no task with ID %d", n)
		}
		add(n)
	}
	if len(ids) == 0 {
		return nil, usageError("no tasks in that range")
	}
	return ids, nil
}

func parseRange(s string) (lo, hi int, ok bool) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	lo, err1 := strconv.Atoi(parts[0])
	hi, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || lo < 1 || hi < lo {
		return 0, 0, false
	}
	return lo, hi, true
}

func joinIDs(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = "#" + strconv.Itoa(id)
	}
	return strings.Join(parts, " ")
}

// flagValue reads "--name value" or "--name=value".
func flagValue(args []string, i *int, name string) string {
	a := args[*i]
	if strings.HasPrefix(a, name+"=") {
		return a[len(name)+1:]
	}
	if *i+1 < len(args) {
		*i++
		return args[*i]
	}
	return ""
}

func cmdAdd(s *Store, w io.Writer, args []string, now time.Time) error {
	if len(args) == 0 {
		return usageError(`add needs task text — e.g. tod add "Buy milk" !high due:tomorrow`)
	}
	p, err := parseInput(args, now)
	if err != nil {
		return err
	}
	if p.text == "" {
		return usageError("add needs task text (metadata alone isn't a task)")
	}
	t := Task{
		Text:     p.text,
		Priority: p.pri,
		Tags:     p.tags,
		Project:  p.project,
		Due:      p.due,
		Every:    p.every,
		Created:  now,
	}
	s.PushUndo(fmt.Sprintf("add #%d", s.nextID()))
	t = s.Add(t)
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(w, "%s Added %s  %s\n", glyphCheck(), stID.Render(fmt.Sprintf("#%d", t.ID)), t.Text)
	if meta := metaSummary(t, now); meta != "" {
		fmt.Fprintf(w, "   %s\n", meta)
	}
	return nil
}

// metaSummary is a one-line dim summary of a task's metadata.
func metaSummary(t Task, now time.Time) string {
	var parts []string
	if t.Priority != PriNone {
		parts = append(parts, t.Priority.String())
	}
	for _, tag := range t.Tags {
		parts = append(parts, "@"+tag)
	}
	if t.Project != "" {
		parts = append(parts, "#"+t.Project)
	}
	if d, ok := t.DueDate(); ok {
		parts = append(parts, "due "+strings.ToLower(d.Format("Mon Jan 2")))
	}
	if t.Every != "" {
		parts = append(parts, "repeats "+everyLabel(t.Every))
	}
	return stDim.Render(strings.Join(parts, " · "))
}

func cmdList(s *Store, w io.Writer, args []string, now time.Time) error {
	o := listOpts{now: now, width: termWidth()}
	var query []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-a" || a == "--all":
			o.all = true
		case a == "-d" || a == "--done":
			o.doneOnly = true
		case a == "--plain":
			setPlain()
		case a == "--no-color":
			setNoColor()
		case a == "--tag" || strings.HasPrefix(a, "--tag="):
			o.tag = strings.TrimPrefix(strings.ToLower(flagValue(args, &i, "--tag")), "@")
		case a == "--project" || strings.HasPrefix(a, "--project="):
			o.project = strings.TrimPrefix(strings.ToLower(flagValue(args, &i, "--project")), "#")
		case a == "--pri" || strings.HasPrefix(a, "--pri="):
			p, ok := parsePriority(strings.ToLower(flagValue(args, &i, "--pri")))
			if !ok {
				return usageError("unknown priority (use high, medium, or low)")
			}
			o.pri, o.hasPri = p, true
		case strings.HasPrefix(a, "@") && len(a) > 1:
			o.tag = strings.ToLower(a[1:])
		case strings.HasPrefix(a, "#") && len(a) > 1:
			o.project = strings.ToLower(a[1:])
		case strings.HasPrefix(a, "-"):
			return usageError("unknown flag %s (see `tod help`)", a)
		default:
			query = append(query, a)
		}
	}
	o.query = strings.Join(query, " ")
	fmt.Fprint(w, renderList(s.Tasks(), o))
	return nil
}

func cmdDone(s *Store, w io.Writer, args []string, now time.Time) error {
	if len(args) == 0 {
		return usageError("done needs task IDs — e.g. tod done 1 2  or  tod done 1-3")
	}
	ids, err := parseIDs(s, args)
	if err != nil {
		return err
	}
	s.PushUndo("done " + joinIDs(ids))
	var spawned []Task
	for _, id := range ids {
		t := s.ByID(id)
		if t.Done {
			fmt.Fprintf(w, "%s\n", stDim.Render(fmt.Sprintf("#%d was already done", id)))
			continue
		}
		t.Done = true
		c := now
		t.Completed = &c
		fmt.Fprintf(w, "%s %s  %s\n", glyphCheck(), stID.Render(fmt.Sprintf("#%d", id)), stDoneText.Render(t.Text))
		if t.Every != "" {
			due, _ := t.DueDate()
			spawned = append(spawned, Task{
				Text:     t.Text,
				Priority: t.Priority,
				Tags:     append([]string{}, t.Tags...),
				Project:  t.Project,
				Due:      nextOccurrence(due, t.Every, now).Format("2006-01-02"),
				Every:    t.Every,
				Created:  now,
			})
		}
	}
	for _, nt := range spawned {
		nt = s.Add(nt)
		d, _ := nt.DueDate()
		fmt.Fprintf(w, "   %s\n", stDim.Render(fmt.Sprintf("%s next: #%d due %s", glyphRecur(), nt.ID, strings.ToLower(d.Format("Mon Jan 2")))))
	}
	return s.Save()
}

func cmdUndone(s *Store, w io.Writer, args []string, now time.Time) error {
	if len(args) == 0 {
		return usageError("undone needs task IDs — e.g. tod undone 2")
	}
	ids, err := parseIDs(s, args)
	if err != nil {
		return err
	}
	s.PushUndo("undone " + joinIDs(ids))
	for _, id := range ids {
		t := s.ByID(id)
		if !t.Done {
			fmt.Fprintf(w, "%s\n", stDim.Render(fmt.Sprintf("#%d is not done", id)))
			continue
		}
		t.Done = false
		t.Completed = nil
		fmt.Fprintf(w, "%s Reopened %s  %s\n", glyphCheck(), stID.Render(fmt.Sprintf("#%d", id)), t.Text)
	}
	return s.Save()
}

func cmdRm(s *Store, w io.Writer, args []string, now time.Time) error {
	if len(args) == 0 {
		return usageError("rm needs task IDs — e.g. tod rm 2 4")
	}
	ids, err := parseIDs(s, args)
	if err != nil {
		return err
	}
	s.PushUndo("rm " + joinIDs(ids))
	for _, id := range ids {
		t := s.ByID(id)
		if t == nil {
			continue
		}
		text := t.Text
		s.Delete(id)
		fmt.Fprintf(w, "%s\n", stDim.Render(fmt.Sprintf("Removed #%d  %s", id, text)))
	}
	return s.Save()
}

func cmdEdit(s *Store, w io.Writer, args []string, now time.Time) error {
	if len(args) < 2 {
		return usageError(`edit needs an ID and changes — e.g. tod edit 2 "Buy oat milk" !med due:fri`)
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		return usageError("%q is not a task ID", args[0])
	}
	t := s.ByID(n)
	if t == nil {
		return fmt.Errorf("no task with ID %d", n)
	}
	p, err := parseInput(args[1:], now)
	if err != nil {
		return err
	}
	if p.text == "" && !p.hasPri && !p.hasTags && !p.hasProj && !p.hasDue && !p.hasEvery {
		return usageError("nothing to change — give new text or metadata")
	}
	s.PushUndo(fmt.Sprintf("edit #%d", n))
	if p.text != "" {
		t.Text = p.text
	}
	if p.hasPri {
		t.Priority = p.pri
	}
	if p.hasTags {
		t.Tags = p.tags
	}
	if p.hasProj {
		t.Project = p.project
	}
	if p.hasDue {
		t.Due = p.due
	}
	if p.hasEvery {
		t.Every = p.every
	}
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(w, "%s Updated %s  %s\n", glyphCheck(), stID.Render(fmt.Sprintf("#%d", n)), t.Text)
	if meta := metaSummary(*t, now); meta != "" {
		fmt.Fprintf(w, "   %s\n", meta)
	}
	return nil
}

func cmdPri(s *Store, w io.Writer, args []string, now time.Time) error {
	if len(args) < 2 {
		return usageError("pri needs an ID and a level — e.g. tod pri 2 high")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		return usageError("%q is not a task ID", args[0])
	}
	t := s.ByID(n)
	if t == nil {
		return fmt.Errorf("no task with ID %d", n)
	}
	p, ok := parsePriority(strings.ToLower(args[1]))
	if !ok {
		return usageError("unknown priority %q (use high, medium, low, or none)", args[1])
	}
	s.PushUndo(fmt.Sprintf("pri #%d", n))
	t.Priority = p
	if err := s.Save(); err != nil {
		return err
	}
	label := p.String()
	if p == PriNone {
		label = "none"
	}
	fmt.Fprintf(w, "%s %s priority → %s\n", glyphCheck(), stID.Render(fmt.Sprintf("#%d", n)), label)
	return nil
}

func cmdDue(s *Store, w io.Writer, args []string, now time.Time) error {
	if len(args) < 2 {
		return usageError("due needs an ID and a date — e.g. tod due 2 friday  (or `none` to clear)")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		return usageError("%q is not a task ID", args[0])
	}
	t := s.ByID(n)
	if t == nil {
		return fmt.Errorf("no task with ID %d", n)
	}
	s.PushUndo(fmt.Sprintf("due #%d", n))
	v := strings.ToLower(args[1])
	if v == "none" || v == "-" || v == "" {
		t.Due = ""
	} else {
		d, err := parseDate(v, now)
		if err != nil {
			return err
		}
		t.Due = d.Format("2006-01-02")
	}
	if err := s.Save(); err != nil {
		return err
	}
	if t.Due == "" {
		fmt.Fprintf(w, "%s %s due date cleared\n", glyphCheck(), stID.Render(fmt.Sprintf("#%d", n)))
	} else {
		d, _ := t.DueDate()
		fmt.Fprintf(w, "%s %s due → %s\n", glyphCheck(), stID.Render(fmt.Sprintf("#%d", n)), strings.ToLower(d.Format("Mon Jan 2")))
	}
	return nil
}

func cmdStats(s *Store, w io.Writer, now time.Time) error {
	fmt.Fprint(w, renderStats(s.Tasks(), now))
	return nil
}

func cmdUndo(s *Store, w io.Writer) error {
	desc, ok := s.Undo()
	if !ok {
		fmt.Fprintln(w, stDim.Render("Nothing to undo."))
		return nil
	}
	fmt.Fprintf(w, "%s Undid: %s\n", glyphCheck(), desc)
	return nil
}

func cmdRedo(s *Store, w io.Writer) error {
	desc, ok := s.Redo()
	if !ok {
		fmt.Fprintln(w, stDim.Render("Nothing to redo."))
		return nil
	}
	fmt.Fprintf(w, "%s Redid: %s\n", glyphCheck(), desc)
	return nil
}

func cmdClear(s *Store, w io.Writer, args []string, stdin *os.File) error {
	force := false
	for _, a := range args {
		if a == "-f" || a == "--force" {
			force = true
		} else {
			return usageError("unknown flag %s (clear only supports --force)", a)
		}
	}
	var ids []int
	for _, t := range s.Tasks() {
		if t.Done {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		fmt.Fprintln(w, stDim.Render("Nothing to clear — no completed tasks."))
		return nil
	}
	if !force && term.IsTerminal(int(stdin.Fd())) {
		fmt.Fprintf(w, "Remove %s? [y/N] ", plural(len(ids), "completed task"))
		line, _ := bufio.NewReader(stdin).ReadString('\n')
		if strings.ToLower(strings.TrimSpace(line)) != "y" {
			fmt.Fprintln(w, stDim.Render("Aborted."))
			return nil
		}
	}
	s.PushUndo("clear " + plural(len(ids), "completed task"))
	for _, id := range ids {
		s.Delete(id)
	}
	if err := s.Save(); err != nil {
		return err
	}
	fmt.Fprintf(w, "%s Cleared %s.\n", glyphCheck(), plural(len(ids), "completed task"))
	return nil
}

func cmdExport(s *Store, w io.Writer) error {
	fd := fileData{Version: dataVersion, Tasks: s.Tasks()}
	if fd.Tasks == nil {
		fd.Tasks = []Task{}
	}
	b, err := json.MarshalIndent(fd, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

func cmdPath(s *Store, w io.Writer) {
	fmt.Fprintf(w, "  Data dir   %s\n", s.dir)
	fmt.Fprintf(w, "  Tasks      %s\n", filepath.Join(s.dir, "tasks.json"))
	fmt.Fprintf(w, "  History    %s\n", filepath.Join(s.dir, "undo.json"))
	fmt.Fprintf(w, "  %s stored\n", plural(len(s.Tasks()), "task"))
}

func cmdCompletion(w io.Writer, args []string) error {
	if len(args) == 0 {
		return usageError("completion needs a shell — e.g. tod completion bash")
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(w, bashCompletion)
	case "zsh":
		fmt.Fprint(w, zshCompletion)
	case "fish":
		fmt.Fprint(w, fishCompletion)
	default:
		return usageError("unknown shell %q (use bash, zsh, or fish)", args[0])
	}
	return nil
}
