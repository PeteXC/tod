package main

import (
	"fmt"
	"strings"
	"time"
)

// parsedInput is the result of parsing free-form task input with inline
// metadata, e.g. `tod add Buy milk !high @groceries #home due:fri every:week`.
type parsedInput struct {
	text     string
	pri      Priority
	tags     []string
	project  string
	due      string // ISO date or ""
	every    string
	hasPri   bool
	hasTags  bool
	hasProj  bool
	hasDue   bool
	hasEvery bool
}

// parseInput splits args into words and extracts metadata tokens:
//
//	!high !med !low   priority (also !!! !! ! and !none)
//	@tag              tag (repeatable)
//	#project          project (last one wins)
//	due:<when>        due date (due:none clears)
//	every:<span>      recurrence (every:none clears)
//
// Everything else becomes the task text.
func parseInput(args []string, now time.Time) (parsedInput, error) {
	var p parsedInput
	var words []string
	for _, arg := range args {
		for _, tok := range strings.Fields(arg) {
			switch {
			case strings.HasPrefix(tok, "due:"):
				p.hasDue = true
				v := tok[len("due:"):]
				if v == "" || v == "none" || v == "-" {
					p.due = ""
					continue
				}
				d, err := parseDate(v, now)
				if err != nil {
					return p, err
				}
				p.due = d.Format("2006-01-02")
			case strings.HasPrefix(tok, "every:"):
				p.hasEvery = true
				v := tok[len("every:"):]
				if v == "" || v == "none" || v == "-" {
					p.every = ""
					continue
				}
				e, err := parseEvery(v)
				if err != nil {
					return p, err
				}
				p.every = e
			case strings.HasPrefix(tok, "!"):
				p.hasPri = true
				if isAllBangs(tok) {
					switch {
					case len(tok) == 1:
						p.pri = PriLow
					case len(tok) == 2:
						p.pri = PriMed
					default:
						p.pri = PriHigh
					}
					continue
				}
				pr, ok := parsePriority(strings.ToLower(tok[1:]))
				if !ok {
					return p, fmt.Errorf("unknown priority %q (use !low, !med, !high, or ! !! !!!)", tok)
				}
				p.pri = pr
			case strings.HasPrefix(tok, "@") && len(tok) > 1:
				p.hasTags = true
				p.tags = append(p.tags, strings.ToLower(tok[1:]))
			case strings.HasPrefix(tok, "#") && len(tok) > 1:
				p.hasProj = true
				p.project = strings.ToLower(tok[1:])
			default:
				words = append(words, tok)
			}
		}
	}
	p.text = strings.Join(words, " ")
	return p, nil
}

func isAllBangs(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != '!' {
			return false
		}
	}
	return true
}
