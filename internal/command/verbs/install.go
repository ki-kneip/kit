package verbs

import (
	"context"

	"github.com/ki-kneip/kit/internal/command"
	"github.com/ki-kneip/kit/internal/selfinstall"
)

func init() {
	command.Register("install", Install{})
	command.Register("uninstall", Uninstall{})
}

type Install struct {
}

func (cmd Install) Help() string {
	return `copy kit into %LocalAppData%\Programs\kit and add it to your user PATH
usage:
  kit install     no admin rights needed; open a new terminal afterwards
notes:
  Windows only for now — elsewhere, use ` + "`go install ...cmd/kit@latest`" + ` or your package manager.`
}

func (cmd Install) Execute(ctx context.Context, args ...string) error {
	target, added, err := selfinstall.Install()
	if err != nil {
		return err
	}
	command.Line(0, command.Title("installed"), command.Path(target))
	if added {
		command.Line(1, command.Dim("added to PATH — open a new terminal to run `kit` from anywhere"))
	} else {
		command.Line(1, command.Dim("already on PATH"))
	}
	return nil
}

type Uninstall struct {
}

func (cmd Uninstall) Help() string {
	return `remove kit from your user PATH and delete the installed copy`
}

func (cmd Uninstall) Execute(ctx context.Context, args ...string) error {
	dir, err := selfinstall.Uninstall()
	if err != nil {
		return err
	}
	command.Line(0, command.Title("uninstalled"), command.Dim("removed from PATH: "+dir))
	return nil
}
