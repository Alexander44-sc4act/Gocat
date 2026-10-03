//go:build windows

package network

import (
	"fmt"
	"net"
	"time"
)

// SCTPConfig holds SCTP-specific configuration.
type SCTPConfig struct {
	Streams        int
	MaxAttempts    int
	MaxInitTimeout time.Duration
	Heartbeat      bool
	Nodelay        bool
	AutoClose      time.Duration
}

// DefaultSCTPConfig returns SCTP defaults.
func DefaultSCTPConfig() *SCTPConfig {
	return &SCTPConfig{
		Streams:        10,
		MaxAttempts:    4,
		MaxInitTimeout: 60 * time.Second,
		Heartbeat:      true,
	}
}

// SCTPAddr represents an SCTP address.
type SCTPAddr struct {
	IPs  []net.IP
	Port int
}

func (a *SCTPAddr) Network() string { return "sctp" }

func (a *SCTPAddr) String() string {
	if len(a.IPs) == 0 {
		return fmt.Sprintf(":%d", a.Port)
	}
	return net.JoinHostPort(a.IPs[0].String(), fmt.Sprintf("%d", a.Port))
}

// SCTPConn is a placeholder on Windows, where native SCTP is unsupported here.
type SCTPConn struct{}

func (c *SCTPConn) Read(b []byte) (int, error)         { return 0, errSCTPUnsupported() }
func (c *SCTPConn) Write(b []byte) (int, error)        { return 0, errSCTPUnsupported() }
func (c *SCTPConn) Close() error                       { return nil }
func (c *SCTPConn) LocalAddr() net.Addr                { return nil }
func (c *SCTPConn) RemoteAddr() net.Addr               { return nil }
func (c *SCTPConn) SetDeadline(t time.Time) error      { return errSCTPUnsupported() }
func (c *SCTPConn) SetReadDeadline(t time.Time) error  { return errSCTPUnsupported() }
func (c *SCTPConn) SetWriteDeadline(t time.Time) error { return errSCTPUnsupported() }

// SCTPListener is a placeholder on Windows.
type SCTPListener struct{}

func (l *SCTPListener) Accept() (net.Conn, error) { return nil, errSCTPUnsupported() }
func (l *SCTPListener) Close() error              { return nil }
func (l *SCTPListener) Addr() net.Addr            { return nil }

// SCTPInfo contains SCTP connection information.
type SCTPInfo struct {
	State             string
	LocalAddr         *SCTPAddr
	RemoteAddr        *SCTPAddr
	InboundStreams    int
	OutboundStreams   int
	MaxInboundStreams int
	MaxAttempts       int
}

func (info *SCTPInfo) String() string {
	return "SCTP unsupported on Windows"
}

func ResolveSCTPAddr(network, address string) (*SCTPAddr, error) {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	port, err := net.LookupPort("sctp", portText)
	if err != nil {
		return nil, err
	}
	var ips []net.IP
	if host != "" {
		ips, err = net.LookupIP(host)
		if err != nil {
			return nil, err
		}
	}
	return &SCTPAddr{IPs: ips, Port: port}, nil
}

func ListenSCTP(network string, laddr *SCTPAddr, config *SCTPConfig) (*SCTPListener, error) {
	return nil, errSCTPUnsupported()
}

func DialSCTP(network string, laddr, raddr *SCTPAddr, config *SCTPConfig) (*SCTPConn, error) {
	return nil, errSCTPUnsupported()
}

func DialSCTPTimeout(network string, laddr, raddr *SCTPAddr, timeout time.Duration, config *SCTPConfig) (*SCTPConn, error) {
	return nil, errSCTPUnsupported()
}

func IsSCTPSupported() bool { return false }

func (c *SCTPConn) GetSCTPInfo() (*SCTPInfo, error) {
	return nil, errSCTPUnsupported()
}

func errSCTPUnsupported() error {
	return fmt.Errorf("SCTP protocol not supported on Windows")
}
