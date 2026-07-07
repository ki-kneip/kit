package kitfile

type Format int

const (
	FormatDSL Format = iota
	FormatYAML
	FormatXML
)

// File is a discovered task file, not yet parsed.
type File struct {
	Path   string // absolute path
	Scope  string // "" for unscoped (.kit), otherwise basename without extension
	Format Format
}
