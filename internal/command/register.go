package command

import "slices"

// Entry is one reserved verb in the register.
type Entry struct {
	Name    string
	Aliases []string
	Exec    Command
}

var registry []Entry

// Register adds a reserved verb. The verbs package calls this from init(),
// so importing it (even blank) is what populates the register.
func Register(name string, exec Command, aliases ...string) {
	registry = append(registry, Entry{Name: name, Aliases: aliases, Exec: exec})
}

// Find resolves a CLI verb (or one of its aliases) to its Command.
// Reserved verbs take precedence over task names, so main calls this
// before task lookup.
func Find(name string) (Command, bool) {
	for _, e := range registry {
		if e.Name == name || slices.Contains(e.Aliases, name) {
			return e.Exec, true
		}
	}
	return nil, false
}

// Entries returns the registered verbs in registration order.
func Entries() []Entry {
	return slices.Clone(registry)
}
