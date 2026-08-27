package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Priority levels. Zero means no explicit priority.
type Priority int

const (
	PriNone Priority = iota
	PriLow
	PriMed
	PriHigh
)

func (p Priority) String() string {
	switch p {
	case PriLow:
		return "low"
	case PriMed:
		return "medium"
	case PriHigh:
		return "high"
	}
	return ""
}

func parsePriority(s string) (Priority, bool) {
	switch s {
	case "1", "l", "lo", "low":
		return PriLow, true
	case "2", "m", "med", "medium":
		return PriMed, true
	case "3", "h", "hi", "high":
		return PriHigh, true
	case "0", "n", "none", "-":
		return PriNone, true
	}
	return PriNone, false
}

// Task is a single to-do item. Due dates are stored as ISO strings
// (YYYY-MM-DD) to keep the JSON data file clean and portable.
type Task struct {
	ID        int        `json:"id"`
	Text      string     `json:"text"`
	Done      bool       `json:"done"`
	Priority  Priority   `json:"priority,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Project   string     `json:"project,omitempty"`
	Due       string     `json:"due,omitempty"`
	Every     string     `json:"every,omitempty"`
	Created   time.Time  `json:"created"`
	Completed *time.Time `json:"completed,omitempty"`
}

// DueDate parses the task's due date in local time.
func (t *Task) DueDate() (time.Time, bool) {
	if t.Due == "" {
		return time.Time{}, false
	}
	d, err := time.ParseInLocation("2006-01-02", t.Due, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return d, true
}

func (t *Task) hasTag(tag string) bool {
	for _, tg := range t.Tags {
		if tg == tag {
			return true
		}
	}
	return false
}

const dataVersion = 1

type fileData struct {
	Version int    `json:"version"`
	Tasks   []Task `json:"tasks"`
}

// Store holds all tasks and persists them to a JSON file with atomic writes.
type Store struct {
	dir   string
	tasks []Task
}

// dataDir resolves where tod keeps its files: $TOD_HOME, else ~/.tod.
func dataDir() string {
	if v := os.Getenv("TOD_HOME"); v != "" {
		return v
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".tod")
	}
	return ".tod"
}

// OpenStore loads the task file, creating the data directory if needed.
// A missing or empty file is a valid, empty store.
func OpenStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create data directory %s: %w", dir, err)
	}
	s := &Store{dir: dir}
	path := filepath.Join(dir, "tasks.json")
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return s, nil
	}
	var fd fileData
	if err := json.Unmarshal(b, &fd); err != nil {
		return nil, fmt.Errorf("data file %s is corrupt: %w", path, err)
	}
	s.tasks = fd.Tasks
	return s, nil
}

// Save persists the store atomically (temp file + rename), so an
// interrupted write can never corrupt the data file.
func (s *Store) Save() error {
	fd := fileData{Version: dataVersion, Tasks: s.tasks}
	if fd.Tasks == nil {
		fd.Tasks = []Task{}
	}
	b, err := json.MarshalIndent(fd, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(s.dir, "tasks.json"), b)
}

func writeFileAtomic(path string, b []byte) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *Store) Tasks() []Task { return s.tasks }

func (s *Store) ByID(id int) *Task {
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			return &s.tasks[i]
		}
	}
	return nil
}

// nextID returns the lowest unused positive ID, keeping IDs short and
// easy to type. Deleted IDs are reused.
func (s *Store) nextID() int {
	used := make(map[int]bool, len(s.tasks))
	max := 0
	for _, t := range s.tasks {
		used[t.ID] = true
		if t.ID > max {
			max = t.ID
		}
	}
	for i := 1; i <= max; i++ {
		if !used[i] {
			return i
		}
	}
	return max + 1
}

func (s *Store) Add(t Task) Task {
	t.ID = s.nextID()
	s.tasks = append(s.tasks, t)
	return t
}

func (s *Store) Delete(id int) bool {
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return true
		}
	}
	return false
}

func cloneTasks(tasks []Task) []Task {
	out := make([]Task, len(tasks))
	for i, t := range tasks {
		out[i] = t
		if t.Tags != nil {
			out[i].Tags = append([]string{}, t.Tags...)
		}
		if t.Completed != nil {
			c := *t.Completed
			out[i].Completed = &c
		}
	}
	return out
}

// sortPending orders tasks for display: earliest due date first
// (undated last), then highest priority, then lowest ID.
func sortPending(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		di, oki := tasks[i].DueDate()
		dj, okj := tasks[j].DueDate()
		if oki != okj {
			return oki
		}
		if oki && !di.Equal(dj) {
			return di.Before(dj)
		}
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority > tasks[j].Priority
		}
		return tasks[i].ID < tasks[j].ID
	})
}

// sortDone orders completed tasks by most recently completed first.
func sortDone(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		ci, cj := tasks[i].Completed, tasks[j].Completed
		if ci == nil || cj == nil {
			if (ci == nil) != (cj == nil) {
				return cj == nil
			}
			return tasks[i].ID > tasks[j].ID
		}
		if !ci.Equal(*cj) {
			return ci.After(*cj)
		}
		return tasks[i].ID > tasks[j].ID
	})
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
