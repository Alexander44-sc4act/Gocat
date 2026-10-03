// Package sniffer provides packet analysis capabilities
package sniffer

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// Analyzer provides packet analysis and statistics
type Analyzer struct {
	connections    map[string]*ConnectionInfo
	topTalkers     map[string]*HostStats
	portStats      map[uint16]*PortStats
	protocolStats  map[PacketType]int64
	bandwidthStats *BandwidthStats
	mu             sync.RWMutex
}

// ConnectionInfo tracks information about a network connection
type ConnectionInfo struct {
	SrcIP       net.IP
	DstIP       net.IP
	SrcPort     uint16
	DstPort     uint16
	Protocol    PacketType
	PacketCount int64
	ByteCount   int64
	FirstSeen   time.Time
	LastSeen    time.Time
	State       ConnectionState
}

// ConnectionState represents TCP connection state
type ConnectionState int

const (
	StateUnknown ConnectionState = iota
	StateSynSent
	StateSynReceived
	StateEstablished
	StateFinWait
	StateClosed
)

func (cs ConnectionState) String() string {
	switch cs {
	case StateSynSent:
		return "SYN_SENT"
	case StateSynReceived:
		return "SYN_RECEIVED"
	case StateEstablished:
		return "ESTABLISHED"
	case StateFinWait:
		return "FIN_WAIT"
	case StateClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

// HostStats tracks statistics for a host
type HostStats struct {
	IP          net.IP
	PacketsSent int64
	PacketsRecv int64
	BytesSent   int64
	BytesRecv   int64
	Connections int
	FirstSeen   time.Time
	LastSeen    time.Time
	TopPorts    map[uint16]int64
}

// PortStats tracks statistics for a port
type PortStats struct {
	Port        uint16
	Protocol    string
	PacketCount int64
	ByteCount   int64
	Connections int
	ServiceName string
}

// BandwidthStats tracks bandwidth usage over time
type BandwidthStats struct {
	Samples       []BandwidthSample
	WindowSize    time.Duration
	MaxSamples    int
	CurrentBytes  int64
	CurrentPkts   int64
	PeakBytesPS   float64
	PeakPacketsPS float64
	mu            sync.Mutex
}

// BandwidthSample represents a bandwidth measurement
type BandwidthSample struct {
	Timestamp time.Time
	Bytes     int64
	Packets   int64
}

// NewAnalyzer creates a new packet analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		connections:   make(map[string]*ConnectionInfo),
		topTalkers:    make(map[string]*HostStats),
		portStats:     make(map[uint16]*PortStats),
		protocolStats: make(map[PacketType]int64),
		bandwidthStats: &BandwidthStats{
			Samples:    make([]BandwidthSample, 0, 60),
			WindowSize: time.Second,
			MaxSamples: 60,
		},
	}
}

// AnalyzePacket analyzes a packet and updates statistics
func (a *Analyzer) AnalyzePacket(packet *Packet) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update protocol stats
	a.protocolStats[packet.Type]++

	// Update connection tracking
	connKey := a.getConnectionKey(packet)
	if conn, exists := a.connections[connKey]; exists {
		conn.PacketCount++
		conn.ByteCount += int64(packet.Length)
		conn.LastSeen = packet.Timestamp
		a.updateConnectionState(conn, packet)
	} else {
		a.connections[connKey] = &ConnectionInfo{
			SrcIP:       packet.SrcIP,
			DstIP:       packet.DstIP,
			SrcPort:     packet.SrcPort,
			DstPort:     packet.DstPort,
			Protocol:    packet.Type,
			PacketCount: 1,
			ByteCount:   int64(packet.Length),
			FirstSeen:   packet.Timestamp,
			LastSeen:    packet.Timestamp,
			State:       StateUnknown,
		}
	}

	// Update host stats
	a.updateHostStats(packet)

	// Update port stats
	a.updatePortStats(packet)

	// Update bandwidth stats
	a.updateBandwidthStats(packet)
}

func (a *Analyzer) getConnectionKey(packet *Packet) string {
	// Create a consistent key regardless of direction
	src := fmt.Sprintf("%s:%d", packet.SrcIP, packet.SrcPort)
	dst := fmt.Sprintf("%s:%d", packet.DstIP, packet.DstPort)
	if src < dst {
		return fmt.Sprintf("%s-%s-%s", src, dst, packet.Type)
	}
	return fmt.Sprintf("%s-%s-%s", dst, src, packet.Type)
}

func (a *Analyzer) updateConnectionState(conn *ConnectionInfo, packet *Packet) {
	if packet.Type != PacketTypeTCP {
		return
	}

	switch {
	case packet.Flags.SYN && !packet.Flags.ACK:
		conn.State = StateSynSent
	case packet.Flags.SYN && packet.Flags.ACK:
		conn.State = StateSynReceived
	case packet.Flags.ACK && conn.State == StateSynReceived:
		conn.State = StateEstablished
	case packet.Flags.FIN:
		conn.State = StateFinWait
	case packet.Flags.RST:
		conn.State = StateClosed
	}
}

func (a *Analyzer) updateHostStats(packet *Packet) {
	srcKey := packet.SrcIP.String()
	dstKey := packet.DstIP.String()

	// Update source host
	if host, exists := a.topTalkers[srcKey]; exists {
		host.PacketsSent++
		host.BytesSent += int64(packet.Length)
		host.LastSeen = packet.Timestamp
		if host.TopPorts == nil {
			host.TopPorts = make(map[uint16]int64)
		}
		host.TopPorts[packet.SrcPort]++
	} else {
		a.topTalkers[srcKey] = &HostStats{
			IP:          packet.SrcIP,
			PacketsSent: 1,
			BytesSent:   int64(packet.Length),
			FirstSeen:   packet.Timestamp,
			LastSeen:    packet.Timestamp,
			TopPorts:    map[uint16]int64{packet.SrcPort: 1},
		}
	}

	// Update destination host
	if host, exists := a.topTalkers[dstKey]; exists {
		host.PacketsRecv++
		host.BytesRecv += int64(packet.Length)
		host.LastSeen = packet.Timestamp
		if host.TopPorts == nil {
			host.TopPorts = make(map[uint16]int64)
		}
		host.TopPorts[packet.DstPort]++
	} else {
		a.topTalkers[dstKey] = &HostStats{
			IP:          packet.DstIP,
			PacketsRecv: 1,
			BytesRecv:   int64(packet.Length),
			FirstSeen:   packet.Timestamp,
			LastSeen:    packet.Timestamp,
			TopPorts:    map[uint16]int64{packet.DstPort: 1},
		}
	}
}

func (a *Analyzer) updatePortStats(packet *Packet) {
	ports := []uint16{packet.SrcPort, packet.DstPort}
	for _, port := range ports {
		if port == 0 {
			continue
		}
		if stats, exists := a.portStats[port]; exists {
			stats.PacketCount++
			stats.ByteCount += int64(packet.Length)
		} else {
			a.portStats[port] = &PortStats{
				Port:        port,
				Protocol:    packet.Type.String(),
				PacketCount: 1,
				ByteCount:   int64(packet.Length),
				ServiceName: GetServiceName(port),
			}
		}
	}
}

func (a *Analyzer) updateBandwidthStats(packet *Packet) {
	a.bandwidthStats.mu.Lock()
	defer a.bandwidthStats.mu.Unlock()

	a.bandwidthStats.CurrentBytes += int64(packet.Length)
	a.bandwidthStats.CurrentPkts++

	// Add sample every second
	now := time.Now()
	if len(a.bandwidthStats.Samples) == 0 ||
		now.Sub(a.bandwidthStats.Samples[len(a.bandwidthStats.Samples)-1].Timestamp) >= a.bandwidthStats.WindowSize {

		sample := BandwidthSample{
			Timestamp: now,
			Bytes:     a.bandwidthStats.CurrentBytes,
			Packets:   a.bandwidthStats.CurrentPkts,
		}
		a.bandwidthStats.Samples = append(a.bandwidthStats.Samples, sample)

		// Update peak values
		bytesPS := float64(a.bandwidthStats.CurrentBytes)
		packetsPS := float64(a.bandwidthStats.CurrentPkts)
		if bytesPS > a.bandwidthStats.PeakBytesPS {
			a.bandwidthStats.PeakBytesPS = bytesPS
		}
		if packetsPS > a.bandwidthStats.PeakPacketsPS {
			a.bandwidthStats.PeakPacketsPS = packetsPS
		}

		// Reset counters
		a.bandwidthStats.CurrentBytes = 0
		a.bandwidthStats.CurrentPkts = 0

		// Trim old samples
		if len(a.bandwidthStats.Samples) > a.bandwidthStats.MaxSamples {
			a.bandwidthStats.Samples = a.bandwidthStats.Samples[1:]
		}
	}
}

// GetTopTalkers returns the top N hosts by traffic
func (a *Analyzer) GetTopTalkers(n int) []*HostStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	hosts := make([]*HostStats, 0, len(a.topTalkers))
	for _, host := range a.topTalkers {
		hosts = append(hosts, host)
	}

	sort.Slice(hosts, func(i, j int) bool {
		totalI := hosts[i].BytesSent + hosts[i].BytesRecv
		totalJ := hosts[j].BytesSent + hosts[j].BytesRecv
		return totalI > totalJ
	})

	if n > len(hosts) {
		n = len(hosts)
	}
	return hosts[:n]
}

// GetTopPorts returns the top N ports by traffic
func (a *Analyzer) GetTopPorts(n int) []*PortStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ports := make([]*PortStats, 0, len(a.portStats))
	for _, port := range a.portStats {
		ports = append(ports, port)
	}

	sort.Slice(ports, func(i, j int) bool {
		return ports[i].ByteCount > ports[j].ByteCount
	})

	if n > len(ports) {
		n = len(ports)
	}
	return ports[:n]
}

// GetActiveConnections returns active connections
func (a *Analyzer) GetActiveConnections() []*ConnectionInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	conns := make([]*ConnectionInfo, 0, len(a.connections))
	for _, conn := range a.connections {
		conns = append(conns, conn)
	}

	sort.Slice(conns, func(i, j int) bool {
		return conns[i].LastSeen.After(conns[j].LastSeen)
	})

	return conns
}

// GetProtocolDistribution returns protocol distribution
func (a *Analyzer) GetProtocolDistribution() map[string]int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	dist := make(map[string]int64)
	for proto, count := range a.protocolStats {
		dist[proto.String()] = count
	}
	return dist
}

// GetBandwidthHistory returns bandwidth history
func (a *Analyzer) GetBandwidthHistory() []BandwidthSample {
	a.bandwidthStats.mu.Lock()
	defer a.bandwidthStats.mu.Unlock()

	samples := make([]BandwidthSample, len(a.bandwidthStats.Samples))
	copy(samples, a.bandwidthStats.Samples)
	return samples
}

// GetServiceName returns the service name for a port
func GetServiceName(port uint16) string {
	services := map[uint16]string{
		20:    "FTP-DATA",
		21:    "FTP",
		22:    "SSH",
		23:    "TELNET",
		25:    "SMTP",
		53:    "DNS",
		67:    "DHCP",
		68:    "DHCP",
		80:    "HTTP",
		110:   "POP3",
		123:   "NTP",
		143:   "IMAP",
		161:   "SNMP",
		443:   "HTTPS",
		445:   "SMB",
		465:   "SMTPS",
		587:   "SUBMISSION",
		993:   "IMAPS",
		995:   "POP3S",
		1433:  "MSSQL",
		1521:  "ORACLE",
		3306:  "MYSQL",
		3389:  "RDP",
		5432:  "POSTGRESQL",
		5900:  "VNC",
		6379:  "REDIS",
		8080:  "HTTP-ALT",
		8443:  "HTTPS-ALT",
		27017: "MONGODB",
	}

	if name, exists := services[port]; exists {
		return name
	}
	return ""
}

// Reset resets all analyzer statistics
func (a *Analyzer) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.connections = make(map[string]*ConnectionInfo)
	a.topTalkers = make(map[string]*HostStats)
	a.portStats = make(map[uint16]*PortStats)
	a.protocolStats = make(map[PacketType]int64)

	a.bandwidthStats.mu.Lock()
	a.bandwidthStats.Samples = make([]BandwidthSample, 0, 60)
	a.bandwidthStats.CurrentBytes = 0
	a.bandwidthStats.CurrentPkts = 0
	a.bandwidthStats.PeakBytesPS = 0
	a.bandwidthStats.PeakPacketsPS = 0
	a.bandwidthStats.mu.Unlock()
}
