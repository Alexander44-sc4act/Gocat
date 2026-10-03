//go:build windows

package terminal

import "os"

// Guard puts stdin into raw mode and returns a restore func. Windows stub:
// restores are best-effort; see guard_unix.go for the full implementation.
func Guard() (restore func(), err error) {
	ts, err := MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return func() {}, err
	}
	once := false
	return func() {
		if !once {
			once = true
			_ = ts.Restore()
		}
	}, nil
}

// RestoreAll restores every terminal currently held by Guard (no-op here;
// the unix implementation tracks guarded states).
func RestoreAll() {}
