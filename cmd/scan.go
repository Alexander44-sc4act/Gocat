package cmd

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/realibrahimsql/Gocat/internal/logger"
	"github.com/spf13/cobra"
)

var (
	scanTimeout   = 3 * time.Second // Default timeout
	concurrency   int
	portRange     string
	verboseOutput bool
	onlyOpen      bool
	useUDPScan    bool
	forceIPv6Scan bool
	forceIPv4Scan bool
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [host] [ports]",
	Short: "Port scanner for network reconnaissance",
	Long: `A fast and efficient port scanner that can scan single ports, port ranges,
or common ports on target hosts. Supports both TCP and UDP scanning with
configurable concurrency and timeout settings.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		ports := "1-1000"
		if len(args) > 1 {
			ports = args[1]
		}

		if portRange != "" {
			ports = portRange
		}

		// Read global flags
		if globalUDP, _ := cmd.Root().PersistentFlags().GetBool("udp"); globalUDP {
			useUDPScan = true
		}
		if globalIPv4, _ := cmd.Root().PersistentFlags().GetBool("ipv4"); globalIPv4 {
			forceIPv4Scan = true
		}
		if globalIPv6, _ := cmd.Root().PersistentFlags().GetBool("ipv6"); globalIPv6 {
			forceIPv6Scan = true
		}
		if globalScanTimeout, _ := cmd.Root().PersistentFlags().GetDuration("scan-timeout"); globalScanTimeout > 0 {
			scanTimeout = globalScanTimeout
		}
		if globalPortRange, _ := cmd.Root().PersistentFlags().GetString("port-range"); globalPortRange != "" && portRange == "" {
			ports = globalPortRange
		}

		logger.Info("Starting port scan on %s for ports %s", host, ports)

		portList, err := parsePortRange(ports)
		if err != nil {
			logger.Error("Invalid port range: %v", err)
			return
		}

		scanPorts(host, portList)
	},
}

func parsePortRange(portStr string) ([]int, error) {
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		return nil, fmt.Errorf("empty port specification")
	}
	var ports []int
	seen := make(map[int]struct{})

	addPort := func(p int, raw string) error {
		if p < 1 || p > 65535 {
			return fmt.Errorf("invalid port %q: must be 1-65535", raw)
		}
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			ports = append(ports, p)
		}
		return nil
	}

	if strings.Contains(portStr, ",") {
		// Handle comma-separated ports: 22,80,443 (ranges allowed per item: 80-90,443)
		for _, part := range strings.Split(portStr, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				return nil, fmt.Errorf("empty port in list %q", portStr)
			}
			if strings.Contains(part, "-") {
				sub, err := expandRange(part)
				if err != nil {
					return nil, err
				}
				for _, p := range sub {
					if err := addPort(p, part); err != nil {
						return nil, err
					}
				}
				continue
			}
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", part)
			}
			if err := addPort(port, part); err != nil {
				return nil, err
			}
		}
	} else if strings.Contains(portStr, "-") {
		sub, err := expandRange(portStr)
		if err != nil {
			return nil, err
		}
		for _, p := range sub {
			if err := addPort(p, portStr); err != nil {
				return nil, err
			}
		}
	} else {
		// Single port
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", portStr)
		}
		if err := addPort(port, portStr); err != nil {
			return nil, err
		}
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("no valid ports in %q", portStr)
	}
	if len(ports) > 65535 {
		return nil, fmt.Errorf("port list too large: %d ports", len(ports))
	}

	return ports, nil
}

func expandRange(portStr string) ([]int, error) {
	parts := strings.Split(portStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid port range format %q: expected start-end", portStr)
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start port: %s", parts[0])
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end port: %s", parts[1])
	}
	if start < 1 || start > 65535 || end < 1 || end > 65535 {
		return nil, fmt.Errorf("port range %q out of bounds: must be 1-65535", portStr)
	}
	if start > end {
		return nil, fmt.Errorf("invalid port range %q: start %d > end %d", portStr, start, end)
	}
	if end-start > 65534 {
		return nil, fmt.Errorf("port range %q too large", portStr)
	}
	ports := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		ports = append(ports, i)
	}
	return ports, nil
}

func scanPorts(host string, ports []int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			isOpen := scanPort(host, p)
			if isOpen || !onlyOpen {
				printResult(host, p, isOpen)
			}
		}(port)
	}

	wg.Wait()
}

func scanPort(host string, port int) bool {
	network := "tcp"
	if useUDPScan {
		return scanUDPPort(host, port)
	}

	if forceIPv6Scan {
		network += "6"
	} else if forceIPv4Scan {
		network += "4"
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout(network, address, scanTimeout)
	if err != nil {
		return false
	}
	if err := conn.Close(); err != nil {
		// Log close error but don't fail the scan
		logger.Warn("Failed to close connection: %v", err)
	}
	return true
}

// scanUDPPort probes a UDP port honestly: UDP Dial alone always succeeds,
// so we must send a probe and wait. Closed ports typically yield ICMP
// port-unreachable (read error); open/filtered ports time out or reply.
func scanUDPPort(host string, port int) bool {
	network := "udp"
	if forceIPv6Scan {
		network += "6"
	} else if forceIPv4Scan {
		network += "4"
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout(network, address, scanTimeout)
	if err != nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()

	// Send a zero-length + generic probe. Some services respond to any datagram.
	_ = conn.SetDeadline(time.Now().Add(scanTimeout))
	if _, err := conn.Write([]byte{0x00}); err != nil {
		// ICMP unreachable often surfaces here on connected UDP sockets.
		return false
	}

	buf := make([]byte, 1024)
	if _, err := conn.Read(buf); err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			// No response: open|filtered. Report open to match TCP semantics
			// (caller prints OPEN); verbose mode distinguishes via --open=false.
			return true
		}
		// ICMP port unreachable / refused -> closed.
		return false
	}
	return true
}

func printResult(host string, port int, isOpen bool) {
	theme := logger.GetCurrentTheme()
	if isOpen {
		if _, err := theme.Success.Printf("[+] %s:%d - OPEN\n", host, port); err != nil {
			log.Printf("Error printing success message: %v", err)
		}
	} else if verboseOutput {
		if _, err := theme.Error.Printf("[-] %s:%d - CLOSED\n", host, port); err != nil {
			log.Printf("Error printing error message: %v", err)
		}
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Scan-specific flags
	scanCmd.Flags().IntVar(&concurrency, "concurrency", 100, "Number of concurrent scans")
	scanCmd.Flags().StringVar(&portRange, "ports", "", "Port range to scan (e.g., 1-1000, 22,80,443)")
	scanCmd.Flags().BoolVar(&verboseOutput, "verbose-scan", false, "Show closed ports as well")
	scanCmd.Flags().BoolVar(&verboseOutput, "scan-verbose", false, "Show closed ports as well (deprecated; use --verbose-scan)")
	scanCmd.Flags().BoolVar(&onlyOpen, "open", true, "Show only open ports")

	// Note: Global flags are used for common options:
	// --udp (global) instead of --scan-udp
	// --ipv4, --ipv6 (global) instead of --scan-ipv4/ipv6
	// --scan-timeout (global, hidden) for port scan timeout
}
