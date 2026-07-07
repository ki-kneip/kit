package kitfile

import (
	"fmt"
	"regexp"
	"strings"
)

// taskHeader matches a task declaration: `name:` optionally followed by a
// description. Only at column 0 — indented lines belong to a task body.
var taskHeader = regexp.MustCompile(`^([A-Za-z0-9_-]+):\s*(.*)$`)

// taskName is the bare name rule, shared with the other frontends.
var taskName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// ParseDSL parses kit's own task format. filename is used in errors only.
//
// The grammar is line-oriented, one construct: a header at column 0 opens a
// task, everything indented below it is the task's shell body, verbatim.
// After the ":" is always a description, never a command.
func ParseDSL(src []byte, filename string) ([]Task, error) {
	var (
		tasks   []Task
		current *Task    // task whose body is being collected, nil at top level
		body    []string // raw body lines of current, original indentation
		seen    = map[string]int{}
	)

	fail := func(n int, format string, args ...any) error {
		return fmt.Errorf("%s:%d: %s", filename, n, fmt.Sprintf(format, args...))
	}

	flush := func() error {
		if current == nil {
			return nil
		}
		run := dedent(body)
		if run == "" {
			return fmt.Errorf("%s: task %q has no commands", filename, current.Name)
		}
		current.Run = run
		tasks = append(tasks, *current)
		current, body = nil, nil
		return nil
	}

	lines := strings.Split(string(src), "\n")
	for i, raw := range lines {
		n := i + 1
		line := strings.TrimRight(raw, "\r") // tolerate CRLF files on Windows

		indented := line != strings.TrimLeft(line, " \t")
		blank := strings.TrimSpace(line) == ""

		// Body lines: anything indented (or blank) while a task is open.
		if current != nil && (indented || blank) {
			body = append(body, line)
			continue
		}

		if blank {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if indented {
			return nil, fail(n, "indented line outside a task")
		}

		// Column 0 and not a comment: must be a task header.
		m := taskHeader.FindStringSubmatch(line)
		if m == nil {
			return nil, fail(n, "expected `name: description`, got %q", line)
		}
		if err := flush(); err != nil {
			return nil, err
		}

		name := m[1]
		if first, dup := seen[name]; dup {
			return nil, fail(n, "task %q already defined at line %d", name, first)
		}
		seen[name] = n
		current = &Task{Name: name, Desc: strings.TrimSpace(m[2])}
	}

	if err := flush(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// dedent strips the common leading whitespace of the non-blank lines and
// trims blank lines at both ends. Interior blank lines are preserved.
func dedent(lines []string) string {
	// trim leading/trailing blank lines
	start, end := 0, len(lines)
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	lines = lines[start:end]
	if len(lines) == 0 {
		return ""
	}

	prefix := leadingWhitespace(lines[0])
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "" {
			continue
		}
		prefix = commonPrefix(prefix, leadingWhitespace(l))
	}

	out := make([]string, len(lines))
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			out[i] = ""
			continue
		}
		out[i] = strings.TrimPrefix(l, prefix)
	}
	return strings.Join(out, "\n")
}

func leadingWhitespace(s string) string {
	return s[:len(s)-len(strings.TrimLeft(s, " \t"))]
}

func commonPrefix(a, b string) string {
	max := min(len(a), len(b))
	for i := 0; i < max; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:max]
}
