// Package sniffer provides advanced network packet capture and analysis capabilities
package sniffer

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// PacketType represents the type of network packet
type PacketType int

const (
	PacketTypeTCP PacketType = iota
	PacketTypeUDP
	PacketTypeICMP
	PacketTypeARP
	PacketTypeDNS
	PacketTypeHTTP
	PacketTypeTLS
	PacketTypeSSH
	PacketTypeOther
)

func (pt PacketType) String() string {
	switch pt {
	case PacketTypeTCP:
		return "TCP"
	case PacketTypeUDP:
		return "UDP"
	case PacketTypeICMP:
		return "ICMP"
	case PacketTypeARP:
		return "ARP"
	case PacketTypeDNS:
		return "DNS"
	case PacketTypeHTTP:
		return "HTTP"
	case PacketTypeTLS:
		return "TLS"
	case PacketTypeSSH:
		return "SSH"
	case PacketTypeOther:
		return "OTHER"
	default:
		return "UNKNOWN"
	}
}

// Packet represents a captured network packet
type Packet struct {
	Timestamp   time.Time
	Type        PacketType
	SrcIP       net.IP
	DstIP       net.IP
	SrcPort     uint16
	DstPort     uint16
	Protocol    string
	Length      int
	Payload     []byte
	Flags       TCPFlags
	Metadata    map[string]interface{}
	CaptureInfo CaptureInfo
}

// TCPFlags represents TCP packet flags
type TCPFlags struct {
	SYN bool
	ACK bool
	FIN bool
	RST bool
	PSH bool
	URG bool
}

func (f TCPFlags) String() string {
	var flags []string
	if f.SYN {
		flags = append(flags, "SYN")
	}
	if f.ACK {
		flags = append(flags, "ACK")
	}
	if f.FIN {
		flags = append(flags, "FIN")
	}
	if f.RST {
		flags = append(flags, "RST")
	}
	if f.PSH {
		flags = append(flags, "PSH")
	}
	if f.URG {
		flags = append(flags, "URG")
	}
	return strings.Join(flags, ",")
}

// CaptureInfo contains metadata about packet capture
type CaptureInfo struct {
	Interface   string
	SnapLength  int
	Truncated   bool
	OriginalLen int
}

// Statistics holds packet capture statistics
type Statistics struct {
	TotalPackets   int64
	TCPPackets     int64
	UDPPackets     int64
	ICMPPackets    int64
	ARPPackets     int64
	OtherPackets   int64
	TotalBytes     int64
	DroppedPackets int64
	ErrorPackets   int64
	StartTime      time.Time
	LastPacketTime time.Time
	PacketsPerSec  float64
	BytesPerSec    float64
}

// Config holds sniffer configuration
type Config struct {
	Interface     string
	Filter        string // BPF filter expression
	Promiscuous   bool
	SnapLength    int32
	Timeout       time.Duration
	PacketCount   int // 0 = unlimited
	BufferSize    int
	Verbose       bool
	DecodePayload bool
	SaveFile      string // pcap output file
	OutputFormat  OutputFormat
}

// OutputFormat defines output format type
type OutputFormat int

const (
	OutputFormatText OutputFormat = iota
	OutputFormatJSON
	OutputFormatHex
	OutputFormatPcap
)

// DefaultConfig returns default sniffer configuration
func DefaultConfig() *Config {
	return &Config{
		Promiscuous:   true,
		SnapLength:    65535,
		Timeout:       30 * time.Second,
		PacketCount:   0,
		BufferSize:    4 * 1024 * 1024, // 4MB
		Verbose:       false,
		DecodePayload: false,
		OutputFormat:  OutputFormatText,
	}
}

// PacketHandler is a callback function for processing packets
type PacketHandler func(packet *Packet)

// Sniffer represents a network packet sniffer
type Sniffer struct {
	config     *Config
	stats      *Statistics
	handlers   []PacketHandler
	ctx        context.Context
	cancel     context.CancelFunc
	running    atomic.Bool
	mu         sync.RWMutex
	statsMu    sync.RWMutex
	outputFile *os.File
}

// NewSniffer creates a new packet sniffer
func NewSniffer(config *Config) *Sniffer {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Sniffer{
		config:   config,
		stats:    &Statistics{StartTime: time.Now()},
		handlers: make([]PacketHandler, 0),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// AddHandler adds a packet handler callback
func (s *Sniffer) AddHandler(handler PacketHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, handler)
}

// GetStatistics returns current capture statistics
func (s *Sniffer) GetStatistics() Statistics {
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()

	stats := *s.stats
	duration := time.Since(s.stats.StartTime).Seconds()
	if duration > 0 {
		stats.PacketsPerSec = float64(s.stats.TotalPackets) / duration
		stats.BytesPerSec = float64(s.stats.TotalBytes) / duration
	}
	return stats
}

// UpdateStats updates packet statistics
func (s *Sniffer) UpdateStats(packet *Packet) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()

	atomic.AddInt64(&s.stats.TotalPackets, 1)
	atomic.AddInt64(&s.stats.TotalBytes, int64(packet.Length))
	s.stats.LastPacketTime = packet.Timestamp

	switch packet.Type {
	case PacketTypeTCP:
		atomic.AddInt64(&s.stats.TCPPackets, 1)
	case PacketTypeUDP:
		atomic.AddInt64(&s.stats.UDPPackets, 1)
	case PacketTypeICMP:
		atomic.AddInt64(&s.stats.ICMPPackets, 1)
	case PacketTypeARP:
		atomic.AddInt64(&s.stats.ARPPackets, 1)
	default:
		atomic.AddInt64(&s.stats.OtherPackets, 1)
	}
}

// Stop stops the packet capture
func (s *Sniffer) Stop() {
	s.cancel()
	s.running.Store(false)

	if s.outputFile != nil {
		s.outputFile.Close()
	}
}

// IsRunning returns whether the sniffer is running
func (s *Sniffer) IsRunning() bool {
	return s.running.Load()
}

// ProcessPacket processes a captured packet
func (s *Sniffer) ProcessPacket(packet *Packet) {
	s.UpdateStats(packet)

	s.mu.RLock()
	handlers := s.handlers
	s.mu.RUnlock()

	for _, handler := range handlers {
		handler(packet)
	}
}

// FormatPacket formats a packet for display
func FormatPacket(packet *Packet, format OutputFormat) string {
	switch format {
	case OutputFormatJSON:
		return formatPacketJSON(packet)
	case OutputFormatHex:
		return formatPacketHex(packet)
	default:
		return formatPacketText(packet)
	}
}

func formatPacketText(packet *Packet) string {
	timestamp := packet.Timestamp.Format("15:04:05.000000")
	flags := ""
	if packet.Type == PacketTypeTCP && packet.Flags.String() != "" {
		flags = fmt.Sprintf(" [%s]", packet.Flags.String())
	}

	return fmt.Sprintf("[%s] %s %s:%d -> %s:%d%s (%d bytes)",
		timestamp,
		packet.Type.String(),
		packet.SrcIP,
		packet.SrcPort,
		packet.DstIP,
		packet.DstPort,
		flags,
		packet.Length,
	)
}

func formatPacketJSON(packet *Packet) string {
	return fmt.Sprintf(`{"timestamp":"%s","type":"%s","src_ip":"%s","src_port":%d,"dst_ip":"%s","dst_port":%d,"length":%d}`,
		packet.Timestamp.Format(time.RFC3339Nano),
		packet.Type.String(),
		packet.SrcIP,
		packet.SrcPort,
		packet.DstIP,
		packet.DstPort,
		packet.Length,
	)
}

func formatPacketHex(packet *Packet) string {
	var sb strings.Builder
	sb.WriteString(formatPacketText(packet))
	sb.WriteString("\n")
	if len(packet.Payload) > 0 {
		sb.WriteString(hex.Dump(packet.Payload))
	}
	return sb.String()
}

// DetectProtocol detects the application layer protocol
func DetectProtocol(payload []byte, srcPort, dstPort uint16) PacketType {
	if len(payload) == 0 {
		return PacketTypeOther
	}

	// DNS detection
	if srcPort == 53 || dstPort == 53 {
		return PacketTypeDNS
	}

	// SSH detection
	if srcPort == 22 || dstPort == 22 {
		return PacketTypeSSH
	}

	// HTTP detection
	httpMethods := []string{"GET ", "POST ", "PUT ", "DELETE ", "HEAD ", "OPTIONS ", "HTTP/"}
	payloadStr := string(payload[:min(100, len(payload))])
	for _, method := range httpMethods {
		if strings.HasPrefix(payloadStr, method) {
			return PacketTypeHTTP
		}
	}

	// TLS detection
	if len(payload) >= 6 && payload[0] == 0x16 && payload[1] == 0x03 {
		return PacketTypeTLS
	}

	return PacketTypeOther
}

// FormatBytes formats bytes to human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
