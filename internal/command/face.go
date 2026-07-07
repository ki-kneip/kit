package command

import "strings"

// The register owns every verb-listing display: the short Face (bare `kit`)
// and the long Overview (`kit help`). Verbs render their own details, but
// anything that enumerates verbs lives here.

// Face prints the short banner shown when kit runs with no arguments.
func Face() {
	Blank()
	Line(1, Logo("⚙ kit"), Dim("— your commands"))
	Blank()
	Line(1, Path("kit <task> [args...]"), Dim("run a task"))
	Line(1, Dim("verbs:"), verbsLine())
	Line(1, Dim("more:"), Verb("kit help"))
	Blank()
}

// Overview prints the long help: usage patterns plus one line per verb.
func Overview() {
	Blank()
	Line(1, Logo("⚙ kit"), Dim("— your commands"))
	Blank()
	Rows(1, [][2]string{
		{"kit <task> [args...]", "run a task ($1..$n inside the script)"},
		{"kit <scope>/<task>", "run a task from a scoped file"},
		{"kit @<scope>/<task>", "run a task from the global dir"},
		{"kit <verb> [args...]", "run a reserved verb (below)"},
	})
	Blank()
	Line(1, Title("verbs"))
	rows := make([][2]string, 0, len(registry))
	for _, e := range registry {
		name := e.Name
		if len(e.Aliases) > 0 {
			name += " " + aliasSuffix(e.Aliases)
		}
		rows = append(rows, [2]string{name, summary(e.Exec.Help())})
	}
	Rows(1, rows)
	Blank()
}

func verbsLine() string {
	parts := make([]string, 0, len(registry))
	for _, e := range registry {
		s := Verb(e.Name)
		if len(e.Aliases) > 0 {
			s += " " + Dim(aliasSuffix(e.Aliases))
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, Dim(" · "))
}

// aliasSuffix renders alternative names as "(ls, l)".
func aliasSuffix(names []string) string {
	return "(" + strings.Join(names, ", ") + ")"
}
