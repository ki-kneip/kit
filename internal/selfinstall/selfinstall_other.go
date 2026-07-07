//go:build !windows

package selfinstall

import "errors"

var errUnsupported = errors.New(
	"kit install is Windows-only for now — on this OS, use " +
		"`go install github.com/ki-kneip/kit/cmd/kit@latest` or your package manager")

func Install() (installedTo string, addedToPath bool, err error) {
	return "", false, errUnsupported
}

func Uninstall() (removedFrom string, err error) {
	return "", errUnsupported
}
