package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ki-kneip/kit/internal/builtins"
	"github.com/ki-kneip/kit/internal/command"
	_ "github.com/ki-kneip/kit/internal/command/verbs" // populates the verb register
	"github.com/ki-kneip/kit/internal/kitfile"
	"github.com/ki-kneip/kit/internal/runner"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z"
// (see .github/workflows/release.yml); "dev" for local/go-install builds.
var version = "dev"

func main() {
	command.Version = version
	os.Exit(run())
}

// run keeps the exit code in one place: os.Exit in main would skip defers.
func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// kit's own flags come before the target: kit -l 3 build
	fs := flag.NewFlagSet("kit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	loop := fs.Int("l", 1, "run the target N times (stops at the first failure)")
	fs.IntVar(loop, "loop", 1, "")
	every := fs.Duration("e", 0, "rerun the target at this interval until Ctrl+C (e.g. 5s)")
	fs.DurationVar(every, "every", 0, "")
	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			command.Overview()
			return 0
		}
		fmt.Fprintln(os.Stderr, "kit:", err)
		return 2
	}
	args := fs.Args()

	if len(args) == 0 {
		command.Face()
		return 0
	}

	// `kit ...` inside scripts runs in-process; see builtins.KitDispatch.
	builtins.KitDispatch = func(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
		if len(args) == 0 {
			return errors.New("nothing to run")
		}
		return dispatch(ctx, args, runner.Options{Stdin: stdin, Stdout: stdout, Stderr: stderr})
	}

	// watch mode: keep going, even through failures, until Ctrl+C
	if *every > 0 {
		for {
			if err := dispatch(ctx, args, runner.Options{}); err != nil {
				fmt.Fprintln(os.Stderr, "kit:", err)
			}
			select {
			case <-ctx.Done():
				return 0
			case <-time.After(*every):
			}
		}
	}

	for i := 0; i < *loop; i++ {
		if err := dispatch(ctx, args, runner.Options{}); err != nil {
			fmt.Fprintln(os.Stderr, "kit:", err)
			return runner.ExitCode(err)
		}
		if ctx.Err() != nil {
			break
		}
	}
	return 0
}

// dispatch resolves and runs one target, in precedence order:
// reserved verb > task > scope listing. Callers own printing the error.
func dispatch(ctx context.Context, args []string, opts runner.Options) error {
	if cmd, ok := command.Find(args[0]); ok {
		return cmd.Execute(ctx, args[1:]...)
	}

	root, target, err := kitfile.RootFor(args[0])
	if err != nil {
		return err
	}

	tasks, err := kitfile.Load(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	scope, name := "", target
	if s, n, ok := strings.Cut(target, "/"); ok {
		scope, name = s, n
	}

	if task, ok := kitfile.FindTask(tasks, scope, name); ok {
		dotenv, err := kitfile.DotEnv(root)
		if err != nil {
			return err
		}
		opts.Params = args[1:]
		if len(dotenv) > 0 {
			// .env entries come last: they win over inherited values
			opts.Env = append(os.Environ(), dotenv...)
		}
		return runner.Run(ctx, task.Run, opts)
	}

	// a bare scope name lists its tasks (discovery for free)
	if scope == "" {
		if scoped := kitfile.ScopeTasks(tasks, target); len(scoped) > 0 {
			label := target
			if kitfile.IsGlobal(args[0]) {
				label = "@" + target
			}
			command.TaskList(label, scoped)
			return nil
		}
	}

	return fmt.Errorf("unknown task or verb %q — try `kit ls` or `kit shell %s`", args[0], strings.Join(args, " "))
}
