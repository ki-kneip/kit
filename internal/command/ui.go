package command

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ui.go centralizes the styles and small helpers every CLI display uses,
// so the output looks the same no matter which verb printed it.

var (
	styleTitle  = lipgloss.NewStyle().Bold(true)
	styleLogo   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	stylePath   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	styleScope  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	styleVerb   = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	styleDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleItalic = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
)

func Title(s string) string  { return styleTitle.Render(s) }
func Logo(s string) string   { return styleLogo.Render(s) }
func Path(s string) string   { return stylePath.Render(s) }
func Scope(s string) string  { return styleScope.Render(s) }
func Verb(s string) string   { return styleVerb.Render(s) }
func Dim(s string) string    { return styleDim.Render(s) }
func Italic(s string) string { return styleItalic.Render(s) }

// Line prints indented content, one level = two spaces.
func Line(indent int, parts ...string) {
	fmt.Println(strings.Repeat("  ", indent) + strings.Join(parts, " "))
}

// Blank prints an empty line.
func Blank() {
	fmt.Println()
}

// Rows prints aligned two-column rows: left column padded to the widest
// entry, right column dimmed. Padding happens before styling so ANSI
// codes don't break the alignment.
func Rows(indent int, rows [][2]string) {
	widest := 0
	for _, r := range rows {
		if len(r[0]) > widest {
			widest = len(r[0])
		}
	}
	for _, r := range rows {
		Line(indent, Path(fmt.Sprintf("%-*s", widest, r[0])), Dim(r[1]))
	}
}

// summary returns the first line of a help text.
func summary(help string) string {
	line, _, _ := strings.Cut(help, "\n")
	return line
}
