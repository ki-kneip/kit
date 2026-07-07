package kitfile

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseYAML parses the YAML task frontend:
//
//	migrate:
//	  desc: run database migrations
//	  run: goose up
//
// Decoded through yaml.Node instead of a map to keep document order.
func ParseYAML(src []byte, filename string) ([]Task, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(src, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	if doc.Kind == 0 || len(doc.Content) == 0 {
		return nil, nil // empty file
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s: top level must be a map of tasks", filename)
	}

	var tasks []Task
	seen := map[string]int{}
	for i := 0; i < len(root.Content); i += 2 {
		key, value := root.Content[i], root.Content[i+1]
		name := key.Value

		if !taskName.MatchString(name) {
			return nil, fmt.Errorf("%s:%d: invalid task name %q", filename, key.Line, name)
		}
		if first, dup := seen[name]; dup {
			return nil, fmt.Errorf("%s:%d: task %q already defined at line %d", filename, key.Line, name, first)
		}
		seen[name] = key.Line

		var body struct {
			Desc string `yaml:"desc"`
			Run  string `yaml:"run"`
		}
		if err := value.Decode(&body); err != nil {
			return nil, fmt.Errorf("%s:%d: task %q: %w", filename, value.Line, name, err)
		}
		if strings.TrimSpace(body.Run) == "" {
			return nil, fmt.Errorf("%s:%d: task %q has no run", filename, key.Line, name)
		}

		tasks = append(tasks, Task{
			Name: name,
			Desc: strings.TrimSpace(body.Desc),
			Run:  strings.TrimRight(body.Run, "\n"),
		})
	}
	return tasks, nil
}
