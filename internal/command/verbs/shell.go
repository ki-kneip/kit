package verbs

import (
	"context"
	"errors"
	"strings"

	"github.com/ki-kneip/kit/internal/command"
	"github.com/ki-kneip/kit/internal/runner"
)

func init() {
	command.Register("shell", Shell{}, "sh")
}

type Shell struct {
}

func (cmd Shell) Help() string {
	return `run a command in kit's embedded POSIX shell (same on every OS)
usage:
  kit shell echo hi              words are joined into one script
  kit shell "mkdir x && ls x"    quote to use operators (&&, |, ;, $VAR)`
}

func (cmd Shell) Execute(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		return errors.New("nothing to run — try `kit shell echo hi`")
	}
	return runner.Run(ctx, strings.Join(args, " "), runner.Options{})
}
