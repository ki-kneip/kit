package kitfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// LocalDir returns the local root: the directory kit is running from.
// Every non-@ lookup resolves against it.
func LocalDir() (string, error) {
	return os.Getwd()
}

// IsGlobal reports whether a CLI target refers to the global dir ("@...").
func IsGlobal(target string) bool {
	return strings.HasPrefix(target, "@")
}

// RootFor resolves a CLI target to the directory it should be looked up
// in, returning the target stripped of its global marker.
func RootFor(target string) (root, rest string, err error) {
	if IsGlobal(target) {
		root, err = GlobalDir()
		return root, strings.TrimPrefix(target, "@"), err
	}
	root, err = LocalDir()
	return root, target, err
}

// GlobalDir returns the global kit directory (AppData\Roaming\kit on
// Windows, ~/.config/kit on Linux), reached from the CLI with "@".
func GlobalDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "kit"), nil
}

func Discover(root string) ([]File, error) {
	files, err := scanDir(root, true)
	if err != nil {
		return nil, err
	}

	sub, err := scanDir(filepath.Join(root, ".kitfiles"), false)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return append(files, sub...), nil
}

// scanDir lists the task files of a single directory, without recursing.
// With onlyKit (project root) just .kit / *.kit are recognized; inside
// .kitfiles, extensionless files and the yaml/xml frontends are accepted too.
func scanDir(path string, onlyKit bool) ([]File, error) {
	dirEntry, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []File
	for _, entry := range dirEntry {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		format, ok := formatOf(name, onlyKit)
		if !ok {
			continue
		}

		files = append(files, File{
			Path:   filepath.Join(path, name),
			Scope:  scopeOf(name),
			Format: format,
		})
	}

	return files, nil
}

// formatOf decides whether name is a task file and, if so, in which format.
func formatOf(name string, onlyKit bool) (Format, bool) {
	ext := filepath.Ext(name)
	if ext == ".kit" {
		return FormatDSL, true
	}
	if onlyKit {
		return 0, false
	}

	switch ext {
	case "":
		return FormatDSL, true
	case ".yaml", ".yml":
		return FormatYAML, true
	case ".xml":
		return FormatXML, true
	}
	return 0, false
}

// scopeOf derives the scope from the file name: ".kit" is unscoped,
// anything else is the base name without its extension.
func scopeOf(name string) string {
	if name == ".kit" {
		return ""
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}
