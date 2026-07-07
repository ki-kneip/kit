package command

import "context"

type Command interface {
	Execute(ctx context.Context, args ...string) error
	// Help returns the verb's help text. The first line must be a short
	// one-line summary — listings (kit help) show only that line.
	Help() string
}
