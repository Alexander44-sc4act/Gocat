package sniffer

import (
	"net"
	"testing"
	"time"
)

func TestNewSniffer(t *testing.T) {
	config := DefaultConfig()
	s := NewSniffer(config)

	if s == nil {
		t.Fatal("NewSniffer returned nil")
	}

	if s.config == nil {
		t.Error("config is nil")
	}

	if s.stats == nil {
		t.Error("stats is nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Promiscuous != true {
		t.Error("Promiscuous should be true by default")
	}

	if config.SnapLength != 65535 {
		t.Errorf("SnapLength = %d, want 65535", config.SnapLength)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", config.Timeout)
	}
}

func TestPacketType(t *testing.T) {
	tests := []struct {
		pt   PacketType
		want string
	}{
		{PacketTypeTCP, "TCP"},
		{PacketTypeUDP, "UDP"},
		{PacketTypeICMP, "ICMP"},
		{PacketTypeARP, "ARP"},
		{PacketTypeDNS, "DNS"},
		{PacketTypeHTTP, "HTTP"},
		{PacketTypeTLS, "TLS"},
		{PacketTypeSSH, "SSH"},
		{PacketTypeOther, "OTHER"},
	}

	for _, tt := range tests {
		if got := tt.pt.String(); got != tt.want {
			t.Errorf("PacketType.String() = %s, want %s", got, tt.want)
		}
	}
}

func TestTCPFlags(t *testing.T) {
	flags := TCPFlags{SYN: true, ACK: true}
	str := flags.String()

	if str != "SYN,ACK" {
		t.Errorf("TCPFlags.String() = %s, want SYN,ACK", str)
	}

	flags2 := TCPFlags{FIN: true, RST: true, PSH: true}
	str2 := flags2.String()

	if str2 != "FIN,RST,PSH" {
		t.Errorf("TCPFlags.String() = %s, want FIN,RST,PSH", str2)
	}
}

func TestDetectProtocol(t *testing.T) {
	tests := []struct {
		payload []byte
		srcPort uint16
		dstPort uint16
		want    PacketType
	}{
		{[]byte("GET / HTTP/1.1"), 12345, 80, PacketTypeHTTP},
		{[]byte("POST /api"), 12345, 8080, PacketTypeHTTP},
		{[]byte{0x16, 0x03, 0x01, 0x00, 0x00, 0x00}, 12345, 443, PacketTypeTLS},
		{[]byte("random data"), 53, 12345, PacketTypeDNS},
		{[]byte("random data"), 12345, 53, PacketTypeDNS},
		{[]byte("random data"), 22, 12345, PacketTypeSSH},
		{[]byte("random data"), 12345, 22, PacketTypeSSH},
		{[]byte("random data"), 12345, 8080, PacketTypeOther},
		{[]byte{}, 12345, 80, PacketTypeOther},
	}

	for _, tt := range tests {
		got := DetectProtocol(tt.payload, tt.srcPort, tt.dstPort)
		if got != tt.want {
			t.Errorf("DetectProtocol(%v, %d, %d) = %v, want %v",
				tt.payload, tt.srcPort, tt.dstPort, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, got, tt.want)
		}
	}
}

func TestStatistics(t *testing.T) {
	s := NewSniffer(nil)

	packet := &Packet{
		Timestamp: time.Now(),
		Type:      PacketTypeTCP,
		SrcIP:     net.ParseIP("192.168.1.1"),
		DstIP:     net.ParseIP("192.168.1.2"),
		SrcPort:   12345,
		DstPort:   80,
		Length:    100,
	}

	s.UpdateStats(packet)
	stats := s.GetStatistics()

	if stats.TotalPackets != 1 {
		t.Errorf("TotalPackets = %d, want 1", stats.TotalPackets)
	}

	if stats.TCPPackets != 1 {
		t.Errorf("TCPPackets = %d, want 1", stats.TCPPackets)
	}

	if stats.TotalBytes != 100 {
		t.Errorf("TotalBytes = %d, want 100", stats.TotalBytes)
	}
}

func TestFormatPacket(t *testing.T) {
	packet := &Packet{
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Type:      PacketTypeTCP,
		SrcIP:     net.ParseIP("192.168.1.1"),
		DstIP:     net.ParseIP("192.168.1.2"),
		SrcPort:   12345,
		DstPort:   80,
		Length:    100,
		Flags:     TCPFlags{SYN: true},
	}

	text := FormatPacket(packet, OutputFormatText)
	if text == "" {
		t.Error("FormatPacket returned empty string")
	}

	json := FormatPacket(packet, OutputFormatJSON)
	if json == "" {
		t.Error("FormatPacket JSON returned empty string")
	}
}
