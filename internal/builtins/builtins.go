// Package builtins implements commands that ship inside the kit binary and
// are injected into the embedded shell via interp.ExecHandlers. They make
// scripts portable: mkdir, rm, cat & friends work the same on Windows,
// where the real executables don't exist.
//
// Builtins win over PATH binaries on purpose — same behavior on every
// machine beats local surprises.
package builtins

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/interp"
)

type builtinFunc func(ctx context.Context, hc interp.HandlerContext, args []string) error

var table = map[string]builtinFunc{
	"mkdir": mkdir,
	"rm":    rm,
	"cp":    cp,
	"mv":    mv,
	"cat":   cat,
	"touch": touch,
	"ls":    ls,
	"hey":   hey,
}

// KitDispatch handles `kit ...` calls inside scripts in-process (no child
// process, Ctrl+C cancels the whole chain). It is set by package main —
// builtins can't import the packages that load and run tasks without an
// import cycle. When nil, `kit` falls through to PATH lookup.
var KitDispatch func(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error

// Middleware intercepts builtin names before the default handler looks
// at PATH. Plugged into the runner with interp.ExecHandlers(Middleware).
func Middleware(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		if len(args) == 0 {
			return next(ctx, args)
		}
		hc := interp.HandlerCtx(ctx)

		if args[0] == "kit" && KitDispatch != nil {
			err := KitDispatch(ctx, args[1:], hc.Stdin, hc.Stdout, hc.Stderr)
			return asStatus(hc, "kit", err)
		}

		fn, ok := table[args[0]]
		if !ok {
			return next(ctx, args)
		}
		return asStatus(hc, args[0], fn(ctx, hc, args[1:]))
	}
}

// asStatus reports a builtin failure the shell way: message on stderr,
// non-zero exit status — never a fatal interpreter error.
func asStatus(hc interp.HandlerContext, name string, err error) error {
	if err == nil {
		return nil
	}
	if status, ok := interp.IsExitStatus(err); ok {
		return interp.ExitStatus(status)
	}
	fmt.Fprintf(hc.Stderr, "%s: %v\n", name, err)
	return interp.ExitStatus(1)
}

// abs resolves p against the script's working directory — relative paths
// must follow the shell's cd, not the kit process's cwd.
func abs(hc interp.HandlerContext, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(hc.Dir, p)
}

// splitFlags separates single-dash flag letters from positional args.
func splitFlags(args []string) (flags map[rune]bool, rest []string) {
	flags = map[rune]bool{}
	for _, a := range args {
		if strings.HasPrefix(a, "-") && len(a) > 1 && a != "--" {
			for _, r := range a[1:] {
				flags[r] = true
			}
			continue
		}
		rest = append(rest, a)
	}
	return flags, rest
}
