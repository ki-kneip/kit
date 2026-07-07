package verbs

import (
	"context"
	"fmt"
	"strings"

	"github.com/ki-kneip/kit/internal/command"
)

func init() {
	command.Register("help", Help{}, "-h", "--help")
}

type Help struct {
}

func (cmd Help) Help() string {
	return `show help for kit or for a single verb
usage:
  kit help          overview and every reserved verb
  kit help <verb>   detailed help for one verb`
}

// Execute with no args prints the register's overview; with an arg,
// that verb's own help text.
func (cmd Help) Execute(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		command.Overview()
		return nil
	}

	verb, ok := command.Find(args[0])
	if !ok {
		return fmt.Errorf("unknown verb %q — run `kit help` to see them all", args[0])
	}
	command.Blank()
	command.Line(1, command.Verb(args[0]))
	for _, l := range strings.Split(verb.Help(), "\n") {
		command.Line(1, l)
	}
	command.Blank()
	return nil
}
