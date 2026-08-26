package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type tuiMode int

const (
	modeNormal tuiMode = iota
	modeAdd
	modeEdit
	modeFilter
)

const footerHelp = "a add · e edit · space done · d del · u undo · / filter · tab all · 1/2/3 pri · t/T due · q quit"

// tuiModel is the interactive UI shown when tod runs with no arguments.
// It reuses the CLI's row rendering so both views look identical.
type tuiModel struct {
	store   *Store
	now     time.Time
	visible []Task
	cursor  int
	showAll bool
	mode    tuiMode
	input   textinput.Model
	filter  string
	width   int
	height  int
	status  string
}

func newTUIModel(s *Store) tuiModel {
	ti := textinput.New()
	ti.CharLimit = 512
	return tuiModel{store: s, now: time.Now(), input: ti, width: 80, height: 24}
}

func (m *tuiModel) refresh() {
	opts := listOpts{all: true, query: m.filter, now: m.now}
	var pend, done []Task
	for _, t := range m.store.Tasks() {
		if !matches(t, opts) {
			continue
		}
		if t.Done {
			done = append(done, t)
		} else {
			pend = append(pend, t)
		}
	}
	sortPending(pend)
	sortDone(done)
	if m.showAll {
		m.visible = append(pend, done...)
	} else {
		m.visible = pend
	}
	if m.cursor > len(m.visible)-1 {
		m.cursor = intMax(0, len(m.visible)-1)
	}
}

func (m tuiModel) Init() tea.Cmd { return nil }

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 14
		return m, nil
	case tea.KeyMsg:
		if m.mode != modeNormal {
			return m.updateInput(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m tuiModel) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		m.input.Blur()
		return m, nil
	case "enter":
		v := strings.TrimSpace(m.input.Value())
		mode := m.mode
		m.mode = modeNormal
		m.input.Blur()
		m.input.SetValue("")
		switch mode {
		case modeAdd:
			m.applyAdd(v)
		case modeEdit:
			m.applyEdit(v)
		case modeFilter:
			m.filter = v
			m.cursor = 0
			m.refresh()
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *tuiModel) applyAdd(v string) {
	if v == "" {
		return
	}
	p, err := parseInput([]string{v}, m.now)
	if err != nil {
		m.status = err.Error()
		return
	}
	if p.text == "" {
		m.status = "task text is required"
		return
	}
	m.store.PushUndo("add (interactive)")
	t := m.store.Add(Task{
		Text:     p.text,
		Priority: p.pri,
		Tags:     p.tags,
		Project:  p.project,
		Due:      p.due,
		Every:    p.every,
		Created:  m.now,
	})
	m.store.Save()
	m.status = fmt.Sprintf("added #%d", t.ID)
	m.refresh()
	for i, vt := range m.visible {
		if vt.ID == t.ID {
			m.cursor = i
		}
	}
}

func (m *tuiModel) applyEdit(v string) {
	if v == "" || len(m.visible) == 0 {
		return
	}
	t := m.store.ByID(m.visible[m.cursor].ID)
	if t == nil {
		return
	}
	p, err := parseInput([]string{v}, m.now)
	if err != nil {
		m.status = err.Error()
		return
	}
	m.store.PushUndo(fmt.Sprintf("edit #%d", t.ID))
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
	m.store.Save()
	m.status = fmt.Sprintf("updated #%d", t.ID)
	m.refresh()
}

func (m tuiModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.filter != "" {
			m.filter = ""
			m.cursor = 0
			m.refresh()
			return m, nil
		}
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = intMax(0, len(m.visible)-1)
	case " ", "enter":
		m.toggleSelected()
	case "a":
		m.mode = modeAdd
		m.input.SetValue("")
		m.input.Prompt = "  add ❯ "
		m.input.Placeholder = "Buy milk !high @groceries due:tomorrow"
		m.status = ""
		return m, m.input.Focus()
	case "e":
		if len(m.visible) == 0 {
			return m, nil
		}
		m.mode = modeEdit
		m.input.SetValue(editPrefill(m.visible[m.cursor]))
		m.input.Prompt = "  edit ❯ "
		m.input.Placeholder = ""
		m.status = ""
		return m, m.input.Focus()
	case "/":
		m.mode = modeFilter
		m.input.SetValue(m.filter)
		m.input.Prompt = "  filter ❯ "
		m.input.Placeholder = "text, @tag, or #project"
		m.status = ""
		return m, m.input.Focus()
	case "d", "x":
		if len(m.visible) == 0 {
			return m, nil
		}
		t := m.visible[m.cursor]
		m.store.PushUndo(fmt.Sprintf("rm #%d", t.ID))
		m.store.Delete(t.ID)
		m.store.Save()
		m.status = fmt.Sprintf("removed #%d", t.ID)
		m.refresh()
	case "u":
		if desc, ok := m.store.Undo(); ok {
			m.status = "undid: " + desc
		} else {
			m.status = "nothing to undo"
		}
		m.refresh()
	case "tab":
		m.showAll = !m.showAll
		m.cursor = 0
		m.refresh()
	case "1", "2", "3", "0":
		if len(m.visible) == 0 {
			return m, nil
		}
		t := m.store.ByID(m.visible[m.cursor].ID)
		if t == nil {
			return m, nil
		}
		var p Priority
		switch msg.String() {
		case "1":
			p = PriLow
		case "2":
			p = PriMed
		case "3":
			p = PriHigh
		}
		m.store.PushUndo(fmt.Sprintf("pri #%d", t.ID))
		t.Priority = p
		m.store.Save()
		if p == PriNone {
			m.status = fmt.Sprintf("#%d priority cleared", t.ID)
		} else {
			m.status = fmt.Sprintf("#%d priority → %s", t.ID, p)
		}
		m.refresh()
	case "t":
		m.setDueSelected(today(m.now))
	case "T":
		m.setDueSelected(today(m.now).AddDate(0, 0, 1))
	}
	return m, nil
}

func (m *tuiModel) toggleSelected() {
	if len(m.visible) == 0 {
		return
	}
	t := m.store.ByID(m.visible[m.cursor].ID)
	if t == nil {
		return
	}
	m.store.PushUndo(fmt.Sprintf("toggle #%d", t.ID))
	if t.Done {
		t.Done = false
		t.Completed = nil
		m.status = fmt.Sprintf("reopened #%d", t.ID)
	} else {
		t.Done = true
		c := m.now
		t.Completed = &c
		m.status = fmt.Sprintf("done #%d", t.ID)
		if t.Every != "" {
			due, _ := t.DueDate()
			nt := m.store.Add(Task{
				Text:     t.Text,
				Priority: t.Priority,
				Tags:     append([]string{}, t.Tags...),
				Project:  t.Project,
				Due:      nextOccurrence(due, t.Every, m.now).Format("2006-01-02"),
				Every:    t.Every,
				Created:  m.now,
			})
			m.status = fmt.Sprintf("done #%d · next is #%d", t.ID, nt.ID)
		}
	}
	m.store.Save()
	m.refresh()
}

func (m *tuiModel) setDueSelected(d time.Time) {
	if len(m.visible) == 0 {
		return
	}
	t := m.store.ByID(m.visible[m.cursor].ID)
	if t == nil {
		return
	}
	m.store.PushUndo(fmt.Sprintf("due #%d", t.ID))
	t.Due = d.Format("2006-01-02")
	m.store.Save()
	m.status = fmt.Sprintf("#%d due %s", t.ID, strings.ToLower(d.Format("Mon Jan 2")))
	m.refresh()
}

// editPrefill reconstructs a task as inline metadata for the edit input.
func editPrefill(t Task) string {
	parts := []string{t.Text}
	switch t.Priority {
	case PriHigh:
		parts = append(parts, "!high")
	case PriMed:
		parts = append(parts, "!med")
	case PriLow:
		parts = append(parts, "!low")
	}
	for _, tag := range t.Tags {
		parts = append(parts, "@"+tag)
	}
	if t.Project != "" {
		parts = append(parts, "#"+t.Project)
	}
	if t.Due != "" {
		parts = append(parts, "due:"+t.Due)
	}
	if t.Every != "" {
		parts = append(parts, "every:"+t.Every)
	}
	return strings.Join(parts, " ")
}

func (m tuiModel) View() string {
	var b strings.Builder

	pendingN := 0
	for _, t := range m.store.Tasks() {
		if !t.Done {
			pendingN++
		}
	}
	header := stHead.Render(" tod") + "  " + stDim.Render(plural(pendingN, "task")+" pending")
	if m.showAll {
		header += stDim.Render(" · showing all")
	}
	if m.filter != "" {
		header += stDim.Render(" · filter: ") + stAccent.Render(m.filter)
	}
	b.WriteString(header + "\n\n")

	avail := m.height - 6
	if avail < 3 {
		avail = 3
	}
	if len(m.visible) == 0 {
		if m.filter != "" {
			b.WriteString(stDim.Render("  No tasks match the filter.") + "\n")
		} else if pendingN == 0 && len(m.store.Tasks()) > 0 {
			b.WriteString("  " + glyphCheck() + " " + stSuccess.Render("All clear — everything is done.") + "\n")
		} else {
			b.WriteString(stDim.Render("  No pending tasks. Press ") + stAccent.Render("a") + stDim.Render(" to add one.") + "\n")
		}
	} else {
		offset := 0
		if m.cursor >= avail {
			offset = m.cursor - avail + 1
		}
		end := offset + avail
		if end > len(m.visible) {
			end = len(m.visible)
		}
		for i := offset; i < end; i++ {
			row := formatTaskRow(m.visible[i], m.now, m.width-2)
			if i == m.cursor {
				row = stAccent.Render("▸") + strings.TrimPrefix(row, " ")
				row = stSelected.Render(pad(row, m.width))
			}
			b.WriteString(row + "\n")
		}
	}

	b.WriteString("\n")
	if m.mode != modeNormal {
		b.WriteString(m.input.View() + "\n")
		b.WriteString("  " + stDim.Render("enter save · esc cancel") + "\n")
	} else {
		if m.status != "" {
			b.WriteString("  " + stSuccess.Render(m.status) + "\n")
		}
		pos := ""
		if len(m.visible) > 0 {
			pos = fmt.Sprintf("  ·  %d/%d", m.cursor+1, len(m.visible))
		}
		b.WriteString("  " + stDim.Render(footerHelp+pos) + "\n")
	}
	return b.String()
}
