package command

import "github.com/ki-kneip/kit/internal/kitfile"

// TaskList prints the tasks of one scope, shown when the CLI gets a bare
// scope name (`kit testes`, `kit @sshinto`).
func TaskList(scope string, tasks []kitfile.Task) {
	Line(0, Title("kit tasks ("+scope+")"))
	rows := make([][2]string, 0, len(tasks))
	for _, t := range tasks {
		desc := t.Desc
		if desc == "" {
			desc = "-"
		}
		rows = append(rows, [2]string{t.Name, desc})
	}
	Rows(1, rows)
	Line(1, Dim("run: kit "+scope+"/<task>"))
}
