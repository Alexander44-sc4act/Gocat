// Helper functions for listener access control and protocol wrappers

package cmd

import (
	"bufio"
	"net"
	"os"
	"strings"

	"github.com/realibrahimsql/Gocat/internal/logger"
)

// Note: Telnet protocol implementation (telnetConn, newTelnetConn) is in connect.go
// and is shared between connect and listen modes

// isIPBlocked checks if an IP is in the deny list
func isIPBlocked(ip string) bool {
	// Load deny list from file if specified
	var denyIPs []string
	denyIPs = append(denyIPs, denyList...)

	if denyFile != "" {
		ips, err := loadIPsFromFile(denyFile)
		if err != nil {
			logger.Warn("Failed to load deny file %q: %v (continuing with %d inline rules)", denyFile, err, len(denyIPs))
		} else {
			denyIPs = append(denyIPs, ips...)
		}
	}

	// Check if IP matches any deny pattern
	for _, pattern := range denyIPs {
		if matchesIPPattern(ip, pattern) {
			return true
		}
	}

	return false
}

// isIPAllowed checks if an IP is in the allow list (returns true if no allow list is set)
func isIPAllowed(ip string) bool {
	// If no allow list is specified, allow all
	if len(allowList) == 0 && allowFile == "" {
		return true
	}

	// Load allow list from file if specified
	var allowIPs []string
	allowIPs = append(allowIPs, allowList...)

	if allowFile != "" {
		ips, err := loadIPsFromFile(allowFile)
		if err != nil {
			logger.Warn("Failed to load allow file %q: %v (continuing with %d inline rules)", allowFile, err, len(allowIPs))
		} else {
			allowIPs = append(allowIPs, ips...)
		}
	}

	// Check if IP matches any allow pattern
	for _, pattern := range allowIPs {
		if matchesIPPattern(ip, pattern) {
			return true
		}
	}

	return false
}

// loadIPsFromFile loads IP addresses from a file (one per line)
func loadIPsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ips []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ips = append(ips, line)
		}
	}

	return ips, scanner.Err()
}

// matchesIPPattern checks if an IP matches a pattern (supports CIDR notation)
func matchesIPPattern(ip, pattern string) bool {
	if ip == pattern {
		return true
	}

	if strings.Contains(pattern, "/") {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		ipAddr := net.ParseIP(ip)
		if ipAddr == nil {
			return false
		}
		return ipNet.Contains(ipAddr)
	}

	return false
}

// Note: crlfWriter is defined in connect.go and shared between connect and listen modes

// applyProtocolWrappers applies telnet and CRLF wrappers if enabled
func applyProtocolWrappers(conn net.Conn) net.Conn {
	finalConn := conn

	if listenTelnetMode {
		logger.Debug("Telnet mode enabled for listener")
		finalConn = newTelnetConn(finalConn)
	}

	// Note: CRLF wrapper is applied in individual handlers (handleNormal, etc.)
	// because it needs to wrap writers, not the connection itself

	return finalConn
}
