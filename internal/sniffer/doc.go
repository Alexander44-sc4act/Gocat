// Package sniffer provides network packet capture and analysis.
//
// # Basic Usage
//
// Create a sniffer with default configuration:
//
//	config := sniffer.DefaultConfig()
//	config.Interface = "eth0"
//	s := sniffer.NewSniffer(config)
//
// Add a packet handler:
//
//	s.AddHandler(func(packet *sniffer.Packet) {
//	    fmt.Println(sniffer.FormatPacket(packet, sniffer.OutputFormatText))
//	})
//
// # Filtering
//
// Use BPF-like filter expressions:
//
//	filter, err := sniffer.NewFilter("tcp port 80")
//	if filter.Match(packet) {
//	    // Process HTTP traffic
//	}
//
// # Analysis
//
// The Analyzer provides detailed traffic analysis:
//
//	analyzer := sniffer.NewAnalyzer()
//	analyzer.AnalyzePacket(packet)
//	topTalkers := analyzer.GetTopTalkers(10)
//	topPorts := analyzer.GetTopPorts(10)
//
// # Statistics
//
// Get capture statistics:
//
//	stats := s.GetStatistics()
//	fmt.Printf("Total packets: %d\n", stats.TotalPackets)
//	fmt.Printf("TCP: %d, UDP: %d\n", stats.TCPPackets, stats.UDPPackets)
package sniffer
