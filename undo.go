package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// undoEntry is a snapshot of the task list before a mutation.
type undoEntry struct {
	Desc  string `json:"desc"`
	Tasks []Task `json:"tasks"`
}

type undoFile struct {
	Undo []undoEntry `json:"undo"`
	Redo []undoEntry `json:"redo"`
}

const undoCap = 50

func (s *Store) undoPath() string { return filepath.Join(s.dir, "undo.json") }

func (s *Store) loadUndo() undoFile {
	var uf undoFile
	b, err := os.ReadFile(s.undoPath())
	if err != nil || len(b) == 0 {
		return uf
	}
	// A corrupt history file is not fatal; it just means no undo.
	json.Unmarshal(b, &uf)
	return uf
}

func (s *Store) saveUndo(uf undoFile) {
	b, err := json.Marshal(uf)
	if err != nil {
		return
	}
	writeFileAtomic(s.undoPath(), b) // best effort
}

// PushUndo snapshots the current state before a mutation. A new change
// invalidates the redo stack.
func (s *Store) PushUndo(desc string) {
	uf := s.loadUndo()
	uf.Undo = append(uf.Undo, undoEntry{Desc: desc, Tasks: cloneTasks(s.tasks)})
	if len(uf.Undo) > undoCap {
		uf.Undo = uf.Undo[len(uf.Undo)-undoCap:]
	}
	uf.Redo = nil
	s.saveUndo(uf)
}

// Undo restores the most recent snapshot, returning its description.
func (s *Store) Undo() (string, bool) {
	uf := s.loadUndo()
	if len(uf.Undo) == 0 {
		return "", false
	}
	e := uf.Undo[len(uf.Undo)-1]
	uf.Undo = uf.Undo[:len(uf.Undo)-1]
	uf.Redo = append(uf.Redo, undoEntry{Desc: e.Desc, Tasks: cloneTasks(s.tasks)})
	s.tasks = e.Tasks
	s.saveUndo(uf)
	s.Save()
	return e.Desc, true
}

// Redo re-applies the most recently undone change.
func (s *Store) Redo() (string, bool) {
	uf := s.loadUndo()
	if len(uf.Redo) == 0 {
		return "", false
	}
	e := uf.Redo[len(uf.Redo)-1]
	uf.Redo = uf.Redo[:len(uf.Redo)-1]
	uf.Undo = append(uf.Undo, undoEntry{Desc: e.Desc, Tasks: cloneTasks(s.tasks)})
	s.tasks = e.Tasks
	s.saveUndo(uf)
	s.Save()
	return e.Desc, true
}
