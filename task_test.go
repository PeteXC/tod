package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := OpenStore(dir)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	return s
}

func TestStoreRoundtrip(t *testing.T) {
	s := tempStore(t)
	when := time.Date(2026, 8, 26, 10, 0, 0, 0, time.Local)
	s.Add(Task{Text: "one", Priority: PriHigh, Tags: []string{"a", "b"}, Project: "x", Due: "2026-08-27", Every: "day", Created: when})
	s.Add(Task{Text: "two", Created: when})
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	s2, err := OpenStore(s.dir)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if len(s2.Tasks()) != 2 {
		t.Fatalf("got %d tasks, want 2", len(s2.Tasks()))
	}
	got := s2.Tasks()[0]
	if got.ID != 1 || got.Text != "one" || got.Priority != PriHigh ||
		len(got.Tags) != 2 || got.Project != "x" || got.Due != "2026-08-27" || got.Every != "day" {
		t.Errorf("roundtrip mismatch: %+v", got)
	}
}

func TestStoreEmptyAndCorrupt(t *testing.T) {
	dir := t.TempDir()
	// Empty file is a valid empty store.
	if err := os.WriteFile(filepath.Join(dir, "tasks.json"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenStore(dir); err != nil {
		t.Errorf("empty file should open cleanly: %v", err)
	}
	// Corrupt file is an error.
	if err := os.WriteFile(filepath.Join(dir, "tasks.json"), []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenStore(dir); err == nil {
		t.Error("corrupt file should error")
	}
}

func TestNextIDReusesGaps(t *testing.T) {
	s := tempStore(t)
	s.Add(Task{Text: "a"})
	s.Add(Task{Text: "b"})
	s.Add(Task{Text: "c"})
	s.Delete(2)
	if got := s.nextID(); got != 2 {
		t.Errorf("nextID after delete = %d, want 2 (lowest unused)", got)
	}
	s.Add(Task{Text: "d"})
	if got := s.nextID(); got != 4 {
		t.Errorf("nextID with no gaps = %d, want 4", got)
	}
}

func TestUndoRedo(t *testing.T) {
	s := tempStore(t)
	s.Add(Task{Text: "a"})

	s.PushUndo("add b")
	s.Add(Task{Text: "b"})
	if len(s.Tasks()) != 2 {
		t.Fatalf("setup: got %d tasks, want 2", len(s.Tasks()))
	}

	desc, ok := s.Undo()
	if !ok || desc != "add b" {
		t.Fatalf("Undo = (%q, %v), want (add b, true)", desc, ok)
	}
	if len(s.Tasks()) != 1 {
		t.Errorf("after undo: got %d tasks, want 1", len(s.Tasks()))
	}

	desc, ok = s.Redo()
	if !ok || desc != "add b" {
		t.Fatalf("Redo = (%q, %v), want (add b, true)", desc, ok)
	}
	if len(s.Tasks()) != 2 {
		t.Errorf("after redo: got %d tasks, want 2", len(s.Tasks()))
	}

	// A new mutation clears the redo stack.
	s.PushUndo("add c")
	s.Add(Task{Text: "c"})
	if _, ok := s.Redo(); ok {
		t.Error("redo should be empty after a new mutation")
	}
}

func TestUndoPersistsAcrossOpens(t *testing.T) {
	s := tempStore(t)
	s.Add(Task{Text: "a"})
	s.PushUndo("add b")
	s.Add(Task{Text: "b"})
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	// A fresh process opening the same dir can still undo.
	s2, err := OpenStore(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s2.Undo(); !ok {
		t.Fatal("undo history should survive reopening the store")
	}
	if len(s2.Tasks()) != 1 {
		t.Errorf("after undo: got %d tasks, want 1", len(s2.Tasks()))
	}
}

func TestSortPending(t *testing.T) {
	tasks := []Task{
		{ID: 1, Text: "no date"},
		{ID: 2, Text: "later", Due: "2026-09-01"},
		{ID: 3, Text: "sooner low", Due: "2026-08-27", Priority: PriLow},
		{ID: 4, Text: "sooner high", Due: "2026-08-27", Priority: PriHigh},
	}
	sortPending(tasks)
	want := []int{4, 3, 2, 1} // due asc, then priority desc, undated last
	for i, id := range want {
		if tasks[i].ID != id {
			t.Errorf("position %d: got #%d, want #%d (order %v)", i, tasks[i].ID, id, want)
		}
	}
}
