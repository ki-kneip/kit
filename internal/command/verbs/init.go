package verbs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ki-kneip/kit/internal/command"
	"github.com/ki-kneip/kit/internal/kitfile"
)

func init() {
	command.Register("init", Init{})
}

const starterKit = `# kit tasks — run one with ` + "`kit <name>`" + `, list files with ` + "`kit ls`" + `.
# Header is "name: description"; the body is plain shell, indented.

hello: say hi from kit
    echo "hello from kit! args: $@"
`

var scopeName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type Init struct {
}

func (cmd Init) Help() string {
	return `create a starter task file
usage:
  kit init                .kit in the current directory
  kit init testes         testes.kit in the current directory
  kit init @              .kit in the global kit directory
  kit init @sshinto       sshinto.kit in the global kit directory
flags:
  -k, --kitfiles          place the file inside .kitfiles/ instead`
}

func (cmd Init) Execute(ctx context.Context, args ...string) error {
	global, scope, inKitfiles, err := parseInitArgs(args)
	if err != nil {
		return err
	}

	var base string
	if global {
		base, err = kitfile.GlobalDir()
	} else {
		base, err = kitfile.LocalDir()
	}
	if err != nil {
		return err
	}

	dir := base
	if inKitfiles {
		dir = filepath.Join(base, ".kitfiles")
	}
	if global || inKitfiles {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	name := ".kit"
	if scope != "" {
		name = scope + ".kit"
	}
	path := filepath.Join(dir, name)

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists in %s", name, dir)
	}

	if err := os.WriteFile(path, []byte(starterKit), 0o644); err != nil {
		return err
	}

	command.Line(0, command.Title("created"), command.Path(path))
	try := "kit hello"
	switch {
	case global && scope != "":
		try = "kit @" + scope + "/hello"
	case global:
		try = "kit @hello"
	case scope != "":
		try = "kit " + scope + "/hello"
	}
	command.Line(1, command.Dim("try: "+try))
	return nil
}

// parseInitArgs reads the target ("", "scope", "@", "@scope") and flags.
func parseInitArgs(args []string) (global bool, scope string, inKitfiles bool, err error) {
	target := ""
	for _, a := range args {
		switch {
		case a == "-k" || a == "--kitfiles":
			inKitfiles = true
		case target == "":
			target = a
		default:
			return false, "", false, fmt.Errorf("unexpected argument %q", a)
		}
	}

	if strings.HasPrefix(target, "@") {
		global = true
		target = strings.TrimPrefix(target, "@")
	}
	if target != "" && !scopeName.MatchString(target) {
		return false, "", false, fmt.Errorf("invalid scope name %q (use letters, digits, - and _)", target)
	}
	return global, target, inKitfiles, nil
}
