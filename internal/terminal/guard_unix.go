//go:build unix

package terminal

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/realibrahimsql/Gocat/internal/logger"
	"golang.org/x/term"
)

var (
	guardMu      sync.Mutex
	guarded      = map[*TerminalState]struct{}{}
	guardSigOnce sync.Once
)

// Guard puts stdin into raw mode and returns a restore func. The restore is:
//   - idempotent (safe to defer and call from signal paths)
//   - registered as a logger exit hook, so Fatal/os.Exit paths restore too
//   - backed by a SIGTERM/SIGHUP watcher that restores every guarded
//     terminal before terminating with the conventional 128+signal status
//
// SIGINT is deliberately not handled here: in raw mode Ctrl+C arrives as a
// byte for the caller to forward, and readline sessions own SIGINT for
// line-cancel semantics. Raw callers without readline must handle SIGINT
// themselves (restore, then return).
func Guard() (restore func(), err error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return func() {}, nil
	}
	ts, err := MakeRaw(fd)
	if err != nil {
		return func() {}, err
	}

	guardMu.Lock()
	guarded[ts] = struct{}{}
	guardMu.Unlock()
	installGuardSignals()

	var once sync.Once
	restore = func() {
		once.Do(func() {
			guardMu.Lock()
			delete(guarded, ts)
			guardMu.Unlock()
			_ = ts.Restore()
		})
	}
	logger.RegisterExitHook(restore)
	return restore, nil
}

// RestoreAll restores every terminal currently held by Guard.
func RestoreAll() {
	guardMu.Lock()
	states := make([]*TerminalState, 0, len(guarded))
	for ts := range guarded {
		states = append(states, ts)
	}
	guarded = map[*TerminalState]struct{}{}
	guardMu.Unlock()
	for _, ts := range states {
		_ = ts.Restore()
	}
}

func installGuardSignals() {
	guardSigOnce.Do(func() {
		ch := make(chan os.Signal, 4)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGHUP)
		go func() {
			for sig := range ch {
				RestoreAll()
				code := 128
				if s, ok := sig.(syscall.Signal); ok {
					code += int(s)
				}
				os.Exit(code)
			}
		}()
	})
}
