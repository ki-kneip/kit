package kitfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// Load discovers and parses every task file under root, filling each
// task's Scope from the file it came from. Files with the same scope are
// merged; two definitions of the same scope/name is an error.
func Load(root string) ([]Task, error) {
	files, err := Discover(root)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	origin := map[string]string{} // scope/name → file that defined it
	for _, f := range files {
		src, err := os.ReadFile(f.Path)
		if err != nil {
			return nil, err
		}

		var parsed []Task
		switch f.Format {
		case FormatDSL:
			parsed, err = ParseDSL(src, filepath.Base(f.Path))
		case FormatYAML:
			parsed, err = ParseYAML(src, filepath.Base(f.Path))
		default:
			continue // XML frontend: only if real demand shows up
		}
		if err != nil {
			return nil, err
		}

		for _, t := range parsed {
			t.Scope = f.Scope
			key := t.Scope + "/" + t.Name
			if prev, dup := origin[key]; dup {
				return nil, fmt.Errorf("task %q defined in both %s and %s",
					t.Name, prev, filepath.Base(f.Path))
			}
			origin[key] = filepath.Base(f.Path)
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

// FindTask looks up a task by scope ("" = unscoped) and name.
func FindTask(tasks []Task, scope, name string) (Task, bool) {
	for _, t := range tasks {
		if t.Scope == scope && t.Name == name {
			return t, true
		}
	}
	return Task{}, false
}

// ScopeTasks returns every task in a scope, for listings.
func ScopeTasks(tasks []Task, scope string) []Task {
	var out []Task
	for _, t := range tasks {
		if t.Scope == scope {
			out = append(out, t)
		}
	}
	return out
}
