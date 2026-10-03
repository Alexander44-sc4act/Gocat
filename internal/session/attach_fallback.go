//go:build !unix

package session

import (
	"fmt"
	"io"
	"os"
)

// InteractSession attaches local stdin to a session on platforms without Unix
// raw terminal support. Ctrl+C behavior is controlled by the host terminal.
func (m *Manager) InteractSession(id int) error {
	sess := m.GetSession(id)
	if sess == nil {
		return fmt.Errorf("session %d not found", id)
	}
	if err := m.AttachSession(id); err != nil {
		return err
	}
	defer m.DetachSession()

	_, err := io.Copy(sess.AsReadWriteCloser(), os.Stdin)
	return err
}
