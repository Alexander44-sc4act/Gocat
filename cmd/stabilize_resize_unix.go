//go:build unix

package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/realibrahimsql/Gocat/internal/logger"
)

func setupStabilizeResize(ptmx *os.File) func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				logger.Debug("Error resizing PTY: %v", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH
	return func() {
		signal.Stop(ch)
		close(ch)
	}
}
