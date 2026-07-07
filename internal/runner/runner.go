package runner

import (
	"context"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"github.com/ki-kneip/kit/internal/builtins"
)

// Options configures how a script is executed.
type Options struct {
	Params []string  // become $1..$n inside the script
	Dir    string    // working directory; empty = current directory
	Env    []string  // KEY=VALUE pairs; nil = inherit os.Environ()
	Stdin  io.Reader // nil = os.Stdin
	Stdout io.Writer // nil = os.Stdout
	Stderr io.Writer // nil = os.Stderr
}

// Run parses and executes a POSIX shell script using the embedded interpreter.
func Run(ctx context.Context, script string, opts Options) error {
	file, err := syntax.NewParser().Parse(strings.NewReader(script), "kitfile")
	if err != nil {
		return err
	}

	env := opts.Env
	if env == nil {
		env = os.Environ()
	}
	stdin, stdout, stderr := opts.Stdin, opts.Stdout, opts.Stderr
	if stdin == nil {
		stdin = os.Stdin
	}
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	runnerOpts := []interp.RunnerOption{
		interp.StdIO(stdin, stdout, stderr),
		interp.Env(expand.ListEnviron(env...)),
		// "--" keeps user args starting with "-" from being read as shell options.
		interp.Params(append([]string{"--"}, opts.Params...)...),
		// kit's built-in commands (portable coreutils & friends) win over PATH.
		interp.ExecHandlers(builtins.Middleware),
	}
	if opts.Dir != "" {
		runnerOpts = append(runnerOpts, interp.Dir(opts.Dir))
	}

	r, err := interp.New(runnerOpts...)
	if err != nil {
		return err
	}
	return r.Run(ctx, file)
}

// ExitCode translates the error returned by Run into a process exit code.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if status, ok := interp.IsExitStatus(err); ok {
		return int(status)
	}
	return 1
}
