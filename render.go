package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// listOpts controls filtering and display for `tod ls`.
type listOpts struct {
	all      bool
	doneOnly bool
	tag      string
	project  string
	pri      Priority
	hasPri   bool
	query    string
	width    int
	now      time.Time
}

func matches(t Task, o listOpts) bool {
	if o.tag != "" && !t.hasTag(o.tag) {
		return false
	}
	if o.project != "" && t.Project != o.project {
		return false
	}
	if o.hasPri && t.Priority != o.pri {
		return false
	}
	if o.query != "" {
		q := strings.ToLower(o.query)
		hay := strings.ToLower(t.Text + " #" + t.Project + " @" + strings.Join(t.Tags, " @"))
		if !strings.Contains(hay, q) {
			return false
		}
	}
	return true
}

// pad right-pads s with spaces to display width w (ANSI-aware).
func pad(s string, w int) string {
	if d := w - lipgloss.Width(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

func truncateTail() string {
	if plain {
		return "..."
	}
	return "…"
}

// dueLabel renders a human, color-coded relative due date.
func dueLabel(t Task, now time.Time) string {
	d, ok := t.DueDate()
	if !ok {
		return ""
	}
	days := int(d.Sub(today(now)).Hours() / 24)
	switch {
	case days < 0:
		if days == -1 {
			return stOverdue.Render("yesterday")
		}
		return stOverdue.Render(fmt.Sprintf("%dd overdue", -days))
	case days == 0:
		return stToday.Render("today")
	case days == 1:
		return stTomorrow.Render("tomorrow")
	case days <= 6:
		return stTomorrow.Render(d.Format("Mon"))
	case days <= 365:
		return stDim.Render(d.Format("Jan 2"))
	default:
		return stDim.Render(d.Format("2006-01-02"))
	}
}

func priBadge(p Priority) string {
	switch p {
	case PriHigh:
		return stHigh.Render("!!!")
	case PriMed:
		return stMed.Render("!!")
	case PriLow:
		return stLow.Render("!")
	}
	return ""
}

// metaParts returns the rendered right-hand metadata for a task row.
func metaParts(t Task, now time.Time) []string {
	var parts []string
	if b := priBadge(t.Priority); b != "" {
		parts = append(parts, b)
	}
	for _, tag := range t.Tags {
		parts = append(parts, stTag.Render("@"+tag))
	}
	if t.Project != "" {
		parts = append(parts, stProject.Render("#"+t.Project))
	}
	if d := dueLabel(t, now); d != "" {
		parts = append(parts, d)
	}
	if t.Every != "" {
		parts = append(parts, stFaint.Render(glyphRecur()+" "+everyLabel(t.Every)))
	}
	return parts
}

// formatTaskRow renders one task line, truncating the text to fit width.
// Shared by the CLI list and the interactive UI so both look identical.
func formatTaskRow(t Task, now time.Time, width int) string {
	id := stID.Render(fmt.Sprintf("%3d", t.ID))
	box := glyphPending()
	if t.Done {
		box = glyphDone()
	}
	meta := strings.Join(metaParts(t, now), " ")

	// Layout: 1 space + 3 id + 2 spaces + box + 2 spaces + text + 2 + meta.
	fixed := 1 + 3 + 2 + 1 + 2
	textW := width - fixed
	if mw := lipgloss.Width(meta); mw > 0 {
		textW -= mw + 2
	}
	if textW < 8 {
		textW = 8
	}
	text := runewidth.Truncate(t.Text, textW, truncateTail())
	if t.Done {
		text = stDoneText.Render(text)
	}
	row := fmt.Sprintf(" %s  %s  %s", id, box, text)
	if meta != "" {
		row += "  " + meta
	}
	return row
}

// bucketFor groups a pending task by due-date urgency.
func bucketFor(t Task, now time.Time) (name string, style lipgloss.Style, rank int) {
	d, ok := t.DueDate()
	if !ok {
		return "SOMEDAY", stDim, 5
	}
	days := int(d.Sub(today(now)).Hours() / 24)
	switch {
	case days < 0:
		return "OVERDUE", stOverdue, 0
	case days == 0:
		return "TODAY", stToday, 1
	case days == 1:
		return "TOMORROW", stTomorrow, 2
	case days <= 6:
		return "THIS WEEK", stHead, 3
	default:
		return "LATER", stDim, 4
	}
}

type bucket struct {
	name  string
	style lipgloss.Style
	tasks []Task
}

func groupPending(pending []Task, now time.Time) []bucket {
	byRank := map[int]*bucket{}
	var order []int
	for _, t := range pending {
		name, style, rank := bucketFor(t, now)
		b, ok := byRank[rank]
		if !ok {
			b = &bucket{name: name, style: style}
			byRank[rank] = b
			order = append(order, rank)
		}
		b.tasks = append(b.tasks, t)
	}
	sort.Ints(order)
	out := make([]bucket, 0, len(order))
	for _, r := range order {
		out = append(out, *byRank[r])
	}
	return out
}

func renderBucketHeader(b *strings.Builder, name string, style lipgloss.Style, n int) {
	b.WriteString("  " + style.Render(name) + " " + stFaint.Render(fmt.Sprintf("· %d", n)) + "\n")
}

// renderList renders the full `tod ls` view: grouped pending tasks,
// an optional completed section, and a progress footer.
func renderList(tasks []Task, o listOpts) string {
	if o.width <= 0 {
		o.width = 80
	}
	var pending, done []Task
	for _, t := range tasks {
		if !matches(t, o) {
			continue
		}
		if t.Done {
			done = append(done, t)
		} else {
			pending = append(pending, t)
		}
	}

	var b strings.Builder
	b.WriteString("\n")

	if len(tasks) == 0 {
		b.WriteString("  " + stDim.Render("Nothing here yet.") + "\n\n")
		b.WriteString("  Add your first task:\n\n")
		b.WriteString("    " + stAccent.Render(`tod add "Learn tod" !high due:today`) + "\n\n")
		return b.String()
	}

	if len(pending)+len(done) == 0 {
		b.WriteString("  " + stDim.Render("No tasks match.") + "\n\n")
		return b.String()
	}

	if o.doneOnly {
		if len(done) == 0 {
			b.WriteString("  " + stDim.Render("No completed tasks yet.") + "\n\n")
			return b.String()
		}
		sortDone(done)
		renderBucketHeader(&b, "COMPLETED", stDoneBox, len(done))
		for _, t := range done {
			b.WriteString(formatTaskRow(t, o.now, o.width) + "\n")
		}
		b.WriteString("\n")
		return b.String()
	}

	sortPending(pending)
	groups := groupPending(pending, o.now)
	if len(groups) == 0 {
		b.WriteString("  " + glyphCheck() + " " + stSuccess.Render(fmt.Sprintf("All clear — %s completed.", plural(len(done), "task"))) + "\n")
	}
	for i, g := range groups {
		if i > 0 {
			b.WriteString("\n")
		}
		renderBucketHeader(&b, g.name, g.style, len(g.tasks))
		for _, t := range g.tasks {
			b.WriteString(formatTaskRow(t, o.now, o.width) + "\n")
		}
	}

	if o.all && len(done) > 0 {
		sortDone(done)
		shown := done
		hidden := 0
		if len(shown) > 10 {
			hidden = len(shown) - 10
			shown = shown[:10]
		}
		b.WriteString("\n")
		renderBucketHeader(&b, "COMPLETED", stDoneBox, len(done))
		for _, t := range shown {
			b.WriteString(formatTaskRow(t, o.now, o.width) + "\n")
		}
		if hidden > 0 {
			b.WriteString("      " + stDim.Render(fmt.Sprintf("… and %d more (tod ls --done)", hidden)) + "\n")
		}
	}

	if line := progressLine(tasks, o); line != "" {
		b.WriteString("\n" + line + "\n")
	}
	return b.String()
}

// progressLine renders the footer progress bar over the filtered task set.
func progressLine(tasks []Task, o listOpts) string {
	total, doneN, overdueN := 0, 0, 0
	base := today(o.now)
	for _, t := range tasks {
		if !matches(t, o) {
			continue
		}
		total++
		if t.Done {
			doneN++
		} else if d, ok := t.DueDate(); ok && d.Before(base) {
			overdueN++
		}
	}
	if total == 0 {
		return ""
	}
	full, empty := barChars()
	const w = 20
	filled := doneN * w / total
	bar := stBarFull.Render(strings.Repeat(full, filled)) + stBarEmpty.Render(strings.Repeat(empty, w-filled))
	s := fmt.Sprintf("  %s  %s", bar, stDim.Render(fmt.Sprintf("%d/%d done · %d%%", doneN, total, doneN*100/total)))
	if overdueN > 0 {
		s += "  " + stOverdue.Render(fmt.Sprintf("%d overdue", overdueN))
	}
	return s
}

func sparkline(counts []int) string {
	max := 0
	for _, c := range counts {
		if c > max {
			max = c
		}
	}
	chars := sparkChars()
	var b strings.Builder
	for _, c := range counts {
		idx := 0
		if max > 0 {
			idx = c * (len(chars) - 1) / max
		}
		b.WriteString(stBarFull.Render(chars[idx]))
	}
	return b.String()
}

// streaks computes the current and best daily completion streaks.
// byDay maps ISO dates to completion counts and only contains days
// with at least one completion.
func streaks(byDay map[string]int, base time.Time) (current, best int) {
	d := base
	if byDay[d.Format("2006-01-02")] == 0 {
		d = d.AddDate(0, 0, -1) // today may simply not have started yet
	}
	for byDay[d.Format("2006-01-02")] > 0 {
		current++
		d = d.AddDate(0, 0, -1)
	}

	if len(byDay) == 0 {
		return current, 0
	}
	days := make([]string, 0, len(byDay))
	for k := range byDay {
		days = append(days, k)
	}
	sort.Strings(days)
	run, prev := 0, ""
	for _, k := range days {
		if prev != "" {
			pd, _ := time.Parse("2006-01-02", prev)
			kd, _ := time.Parse("2006-01-02", k)
			if kd.Sub(pd).Hours() == 24 {
				run++
			} else {
				run = 1
			}
		} else {
			run = 1
		}
		if run > best {
			best = run
		}
		prev = k
	}
	return current, best
}

// renderStats renders the `tod stats` dashboard.
func renderStats(tasks []Task, now time.Time) string {
	var b strings.Builder
	base := today(now)

	pending, doneN, overdueN, doneToday := 0, 0, 0, 0
	byDay := map[string]int{}
	byProject := map[string]int{}
	for _, t := range tasks {
		if t.Done {
			doneN++
			if t.Completed != nil {
				key := t.Completed.Local().Format("2006-01-02")
				byDay[key]++
				if key == base.Format("2006-01-02") {
					doneToday++
				}
			}
			continue
		}
		pending++
		if d, ok := t.DueDate(); ok && d.Before(base) {
			overdueN++
		}
		byProject[t.Project]++
	}

	b.WriteString("\n  " + stHead.Render("Stats") + "\n\n")
	pendLine := fmt.Sprintf("  %-12s %d", "Pending", pending)
	if overdueN > 0 {
		pendLine += "  " + stOverdue.Render(fmt.Sprintf("(%d overdue)", overdueN))
	}
	b.WriteString(pendLine + "\n")
	b.WriteString(fmt.Sprintf("  %-12s %s\n", "Completed",
		stSuccess.Render(fmt.Sprintf("%d", doneN))+stDim.Render(fmt.Sprintf(" total · %d today", doneToday))))

	counts := make([]int, 14)
	maxC := 0
	for i := 0; i < 14; i++ {
		c := byDay[base.AddDate(0, 0, i-13).Format("2006-01-02")]
		counts[i] = c
		if c > maxC {
			maxC = c
		}
	}
	b.WriteString("\n  " + stDim.Render("Last 14 days") + "\n")
	b.WriteString("  " + sparkline(counts) + "\n")
	b.WriteString("  " + stFaint.Render(fmt.Sprintf("%s → today · best %s", base.AddDate(0, 0, -13).Format("Jan 2"), plural(maxC, "task")+"/day")) + "\n")

	current, best := streaks(byDay, base)
	b.WriteString(fmt.Sprintf("\n  %-12s %s\n", "Streak",
		stAccent.Render(plural(current, "day"))+stDim.Render(fmt.Sprintf(" (best %d)", best))))

	if len(byProject) > 0 {
		b.WriteString("\n  " + stDim.Render("Pending by project") + "\n")
		type pc struct {
			label string
			n     int
		}
		var pcs []pc
		for name, n := range byProject {
			label := "#" + name
			if name == "" {
				label = "(no project)"
			}
			pcs = append(pcs, pc{label, n})
		}
		sort.Slice(pcs, func(i, j int) bool {
			if pcs[i].n != pcs[j].n {
				return pcs[i].n > pcs[j].n
			}
			return pcs[i].label < pcs[j].label
		})
		if len(pcs) > 8 {
			pcs = pcs[:8]
		}
		maxN, labelW := 0, 0
		for _, p := range pcs {
			if p.n > maxN {
				maxN = p.n
			}
			if w := runewidth.StringWidth(p.label); w > labelW {
				labelW = w
			}
		}
		full, empty := barChars()
		const w = 12
		for _, p := range pcs {
			filled := 0
			if maxN > 0 {
				filled = p.n * w / maxN
			}
			bar := stBarFull.Render(strings.Repeat(full, filled)) + stBarEmpty.Render(strings.Repeat(empty, w-filled))
			b.WriteString(fmt.Sprintf("  %s  %s %s\n", pad(p.label, labelW), bar, stDim.Render(fmt.Sprintf("%d", p.n))))
		}
	}
	b.WriteString("\n")
	return b.String()
}
