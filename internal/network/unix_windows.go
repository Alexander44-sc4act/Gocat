//go:build windows

package network

import (
	"fmt"
	"net"
	"os"
	"time"
)

type UnixSocketConfig struct {
	Path        string
	Permissions os.FileMode
	Cleanup     bool
	Timeout     time.Duration
}

func DefaultUnixSocketConfig() *UnixSocketConfig {
	return &UnixSocketConfig{Permissions: 0o666, Cleanup: true, Timeout: 30 * time.Second}
}

type UnixDialer struct {
	config *UnixSocketConfig
}

func NewUnixDialer(config *UnixSocketConfig) *UnixDialer {
	if config == nil {
		config = DefaultUnixSocketConfig()
	}
	return &UnixDialer{config: config}
}

func (d *UnixDialer) Dial(socketPath string) (net.Conn, error) {
	return nil, errUnixSocketUnsupported()
}

type UnixListener struct {
	config *UnixSocketConfig
}

func NewUnixListener(config *UnixSocketConfig) *UnixListener {
	if config == nil {
		config = DefaultUnixSocketConfig()
	}
	return &UnixListener{config: config}
}

func (l *UnixListener) Listen(socketPath string) error {
	return errUnixSocketUnsupported()
}

func (l *UnixListener) Accept() (net.Conn, error) {
	return nil, errUnixSocketUnsupported()
}

func (l *UnixListener) Close() error { return nil }

func (l *UnixListener) Addr() net.Addr { return nil }

type UnixSocketInfo struct {
	Path        string
	Permissions os.FileMode
	Size        int64
	ModTime     time.Time
	IsSocket    bool
	InUse       bool
	UID         uint32
	GID         uint32
	Inode       uint64
}

func GetUnixSocketInfo(socketPath string) (*UnixSocketInfo, error) {
	return nil, errUnixSocketUnsupported()
}

func (info *UnixSocketInfo) String() string {
	return "Unix sockets unsupported on Windows"
}

func errUnixSocketUnsupported() error {
	return fmt.Errorf("Unix domain sockets are not supported on Windows")
}
