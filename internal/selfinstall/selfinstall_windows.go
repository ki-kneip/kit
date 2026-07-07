//go:build windows

package selfinstall

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// TargetDir is where kit installs itself: %LocalAppData%\Programs\kit.
// That tree is writable by a normal user, so no elevation is needed.
func TargetDir() (string, error) {
	dir := os.Getenv("LOCALAPPDATA")
	if dir == "" {
		return "", fmt.Errorf("%%LOCALAPPDATA%% is not set")
	}
	return filepath.Join(dir, "Programs", "kit"), nil
}

// Install copies the running executable into TargetDir (unless it's
// already running from there) and adds that directory to the user PATH.
func Install() (installedTo string, addedToPath bool, err error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", false, err
	}
	dir, err := TargetDir()
	if err != nil {
		return "", false, err
	}
	target := filepath.Join(dir, "kit.exe")

	if !samePath(exePath, target) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", false, err
		}
		if err := copyFile(exePath, target); err != nil {
			return "", false, fmt.Errorf("copying kit.exe into %s: %w", dir, err)
		}
	}

	current, err := userPath()
	if err != nil {
		return target, false, fmt.Errorf("reading PATH: %w", err)
	}
	if containsPath(current, dir) {
		return target, false, nil
	}
	if err := setUserPath(joinPath(current, dir)); err != nil {
		return target, false, fmt.Errorf("updating PATH: %w", err)
	}
	broadcastEnvChange()
	return target, true, nil
}

// Uninstall removes TargetDir from the user PATH and deletes the
// installed copy on a best-effort basis (it may be the running exe).
func Uninstall() (removedFrom string, err error) {
	dir, err := TargetDir()
	if err != nil {
		return "", err
	}

	current, err := userPath()
	if err != nil {
		return "", fmt.Errorf("reading PATH: %w", err)
	}
	if containsPath(current, dir) {
		if err := setUserPath(withoutPath(current, dir)); err != nil {
			return "", fmt.Errorf("updating PATH: %w", err)
		}
		broadcastEnvChange()
	}

	_ = os.Remove(filepath.Join(dir, "kit.exe")) // best-effort: may still be running
	return dir, nil
}

func samePath(a, b string) bool {
	ca, errA := filepath.Abs(a)
	cb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(ca), filepath.Clean(cb))
}

func copyFile(src, dst string) error {
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

// --- HKCU\Environment\Path plumbing ---

func userPath() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	val, _, err := k.GetStringValue("Path")
	if err == registry.ErrNotExist {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// setUserPath writes Path as REG_EXPAND_SZ, matching what Windows itself
// uses, so entries like %JAVA_HOME%\bin already in there keep working.
func setUserPath(val string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetExpandStringValue("Path", val)
}

func splitPath(p string) []string {
	var out []string
	for _, s := range strings.Split(p, ";") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func containsPath(pathVar, dir string) bool {
	for _, p := range splitPath(pathVar) {
		if strings.EqualFold(strings.TrimRight(p, `\`), strings.TrimRight(dir, `\`)) {
			return true
		}
	}
	return false
}

func joinPath(pathVar, dir string) string {
	if pathVar == "" {
		return dir
	}
	return strings.TrimRight(pathVar, ";") + ";" + dir
}

func withoutPath(pathVar, dir string) string {
	var kept []string
	for _, p := range splitPath(pathVar) {
		if !strings.EqualFold(strings.TrimRight(p, `\`), strings.TrimRight(dir, `\`)) {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ";")
}

// broadcastEnvChange tells running processes (Explorer, open terminals)
// that the environment changed, the same way the System Properties
// dialog does — without it, PATH only updates after a fresh login.
func broadcastEnvChange() {
	user32 := windows.NewLazySystemDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	const (
		hwndBroadcast   = 0xffff
		wmSettingChange = 0x001A
		smtoAbortIfHung = 0x0002
	)

	env, err := windows.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}
	sendMessageTimeout.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(env)),
		uintptr(smtoAbortIfHung),
		uintptr(5000),
		0,
	)
}
