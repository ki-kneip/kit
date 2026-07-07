// Package selfinstall implements kit's `install`/`uninstall` verbs: kit
// copies itself into a per-user directory and edits the user PATH, with
// no external tool and no admin rights. Windows-only for now — on other
// OSes `go install` or the platform's package manager already do the job.
package selfinstall
