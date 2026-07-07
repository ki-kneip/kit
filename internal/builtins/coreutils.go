package builtins

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"mvdan.cc/sh/v3/interp"
)

// Minimal POSIX-flavored coreutils. Not full clones -- just the everyday
// surface task scripts actually use.

func mkdir(_ context.Context, hc interp.HandlerContext, args []string) error {
	// -p is accepted but implied: MkdirAll is what scripts want anyway.
	_, paths := splitFlags(args)
	if len(paths) == 0 {
		return errors.New("missing operand")
	}
	for _, p := range paths {
		if err := os.MkdirAll(abs(hc, p), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func rm(_ context.Context, hc interp.HandlerContext, args []string) error {
	flags, paths := splitFlags(args)
	recursive := flags['r'] || flags['R']
	force := flags['f']
	if len(paths) == 0 {
		return errors.New("missing operand")
	}
	for _, p := range paths {
		target := abs(hc, p)
		var err error
		if recursive {
			err = os.RemoveAll(target)
		} else {
			err = os.Remove(target)
		}
		if err != nil && !(force && errors.Is(err, os.ErrNotExist)) {
			return err
		}
	}
	return nil
}

func cp(_ context.Context, hc interp.HandlerContext, args []string) error {
	flags, paths := splitFlags(args)
	if len(paths) < 2 {
		return errors.New("usage: cp [-r] source... dest")
	}
	dst := abs(hc, paths[len(paths)-1])
	for _, src := range paths[:len(paths)-1] {
		if err := copyPath(abs(hc, src), dst, flags['r'] || flags['R']); err != nil {
			return err
		}
	}
	return nil
}

func copyPath(src, dst string, recursive bool) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// copying into an existing directory keeps the source's base name
	if di, err := os.Stat(dst); err == nil && di.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	if info.IsDir() {
		if !recursive {
			return fmt.Errorf("%s is a directory (use -r)", src)
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name()), true); err != nil {
				return err
			}
		}
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func mv(_ context.Context, hc interp.HandlerContext, args []string) error {
	_, paths := splitFlags(args)
	if len(paths) != 2 {
		return errors.New("usage: mv source dest")
	}
	src, dst := abs(hc, paths[0]), abs(hc, paths[1])
	if di, err := os.Stat(dst); err == nil && di.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}
	return os.Rename(src, dst)
}

func cat(_ context.Context, hc interp.HandlerContext, args []string) error {
	_, paths := splitFlags(args)
	if len(paths) == 0 {
		_, err := io.Copy(hc.Stdout, hc.Stdin)
		return err
	}
	for _, p := range paths {
		f, err := os.Open(abs(hc, p))
		if err != nil {
			return err
		}
		_, err = io.Copy(hc.Stdout, f)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func touch(_ context.Context, hc interp.HandlerContext, args []string) error {
	_, paths := splitFlags(args)
	if len(paths) == 0 {
		return errors.New("missing operand")
	}
	now := time.Now()
	for _, p := range paths {
		target := abs(hc, p)
		if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			f.Close()
			continue
		}
		if err := os.Chtimes(target, now, now); err != nil {
			return err
		}
	}
	return nil
}

func ls(_ context.Context, hc interp.HandlerContext, args []string) error {
	_, paths := splitFlags(args)
	dir := hc.Dir
	if len(paths) > 0 {
		dir = abs(hc, paths[0])
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintln(hc.Stdout, n)
	}
	return nil
}
