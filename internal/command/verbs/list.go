package verbs

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/ki-kneip/kit/internal/command"
	"github.com/ki-kneip/kit/internal/kitfile"
)

func init() {
	command.Register("list", List{}, "ls")
}

type List struct {
}

func (cmd List) Help() string {
	return `show the task files kit can see
usage:
  kit list      files in the current directory (.kit, *.kit, .kitfiles/)
  kit list @    files in the global kit directory`
}

func formatLabel(f kitfile.Format) string {
	switch f {
	case kitfile.FormatYAML:
		return "yaml"
	case kitfile.FormatXML:
		return "xml"
	default:
		return "dsl"
	}
}

func (cmd List) Execute(ctx context.Context, args ...string) error {
	global := len(args) > 0 && args[0] == "@"

	var root string
	var err error
	if global {
		root, err = kitfile.GlobalDir()
	} else {
		root, err = kitfile.LocalDir()
	}
	if err != nil {
		return err
	}

	files, err := kitfile.Discover(root)
	// A missing global dir just means nothing was registered yet.
	if err != nil && !(global && errors.Is(err, os.ErrNotExist)) {
		return err
	}

	where := "project"
	if global {
		where = "global"
	}
	command.Line(0, command.Title("kit files ("+where+")"), command.Dim(root))

	if len(files) == 0 {
		command.Line(1, command.Italic("no kit files found"))
		return nil
	}

	rows := make([][2]string, 0, len(files))
	for _, f := range files {
		rel, err := filepath.Rel(root, f.Path)
		if err != nil {
			rel = f.Path
		}

		scope := "unscoped"
		if f.Scope != "" {
			scope = f.Scope
			if global {
				scope = "@" + f.Scope
			}
		}
		rows = append(rows, [2]string{filepath.ToSlash(rel), "[" + formatLabel(f.Format) + "] " + scope})
	}
	command.Rows(1, rows)

	return nil
}
