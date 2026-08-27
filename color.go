package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// plain disables color and unicode glyphs (for piping and dumb terminals).
var plain = false

var (
	stID       lipgloss.Style
	stDim      lipgloss.Style
	stFaint    lipgloss.Style
	stDoneBox  lipgloss.Style
	stDoneText lipgloss.Style
	stHigh     lipgloss.Style
	stMed      lipgloss.Style
	stLow      lipgloss.Style
	stTag      lipgloss.Style
	stProject  lipgloss.Style
	stOverdue  lipgloss.Style
	stToday    lipgloss.Style
	stTomorrow lipgloss.Style
	stSuccess  lipgloss.Style
	stError    lipgloss.Style
	stBarFull  lipgloss.Style
	stBarEmpty lipgloss.Style
	stHead     lipgloss.Style
	stAccent   lipgloss.Style
	stSelected lipgloss.Style
)

func init() {
	stID = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	stDim = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	stFaint = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	stDoneBox = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stDoneText = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Strikethrough(true)
	stHigh = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	stMed = lipgloss.NewStyle().Foreground(lipgloss.Color("221"))
	stLow = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	stTag = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	stProject = lipgloss.NewStyle().Foreground(lipgloss.Color("176"))
	stOverdue = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	stToday = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	stTomorrow = lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
	stSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stError = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	stBarFull = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stBarEmpty = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	stHead = lipgloss.NewStyle().Bold(true)
	stAccent = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	stSelected = lipgloss.NewStyle().Background(lipgloss.Color("236"))
}

// setPlain switches to plain output: no color, ASCII-only glyphs.
func setPlain() {
	plain = true
	lipgloss.SetColorProfile(termenv.Ascii)
}

// setNoColor keeps unicode glyphs but strips color.
func setNoColor() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func glyphPending() string {
	if plain {
		return "[ ]"
	}
	return "○"
}

func glyphDone() string {
	if plain {
		return "[x]"
	}
	return stDoneBox.Render("●")
}

func glyphRecur() string {
	if plain {
		return "every"
	}
	return "↻"
}

func glyphCheck() string {
	if plain {
		return "v"
	}
	return stSuccess.Render("✓")
}

func barChars() (full, empty string) {
	if plain {
		return "#", "-"
	}
	return "█", "░"
}

func sparkChars() []string {
	if plain {
		return []string{" ", ".", ":", "-", "=", "+", "*", "#"}
	}
	return []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
}
