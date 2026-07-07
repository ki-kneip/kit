package verbs

import (
	"context"

	"github.com/ki-kneip/kit/internal/command"
)

func init() {
	command.Register("version", Version{})
}

type Version struct {
}

func (cmd Version) Help() string {
	return "print kit's version"
}

func (cmd Version) Execute(ctx context.Context, args ...string) error {
	command.Line(0, command.Verb("kit"), command.Dim(command.Version))
	return nil
}
