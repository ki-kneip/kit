package kitfile

// Task is one runnable unit parsed from a task file.
type Task struct {
	Name  string
	Scope string // filled by the loader from File.Scope, not by parsers
	Desc  string
	Run   string // shell script body, common indentation stripped
}
