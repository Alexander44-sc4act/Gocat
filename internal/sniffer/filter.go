// Package sniffer provides BPF filter parsing and validation
package sniffer

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// Filter represents a packet filter
type Filter struct {
	Expression string
	rules      []FilterRule
	compiled   bool
}

// FilterRule represents a single filter rule
type FilterRule struct {
	Type     FilterType
	Protocol string
	SrcIP    *net.IPNet
	DstIP    *net.IPNet
	SrcPort  uint16
	DstPort  uint16
	PortOp   PortOperator
	Negate   bool
}

// FilterType represents the type of filter
type FilterType int

const (
	FilterTypeProtocol FilterType = iota
	FilterTypeHost
	FilterTypePort
	FilterTypeNet
	FilterTypePortRange
)

// PortOperator represents port comparison operator
type PortOperator int

const (
	PortOpEqual PortOperator = iota
	PortOpLess
	PortOpGreater
	PortOpRange
)

// NewFilter creates a new filter from BPF expression
func NewFilter(expression string) (*Filter, error) {
	f := &Filter{Expression: expression}
	if err := f.compile(); err != nil {
		return nil, err
	}
	return f, nil
}

// compile compiles the BPF expression into filter rules
func (f *Filter) compile() error {
	if f.Expression == "" {
		f.compiled = true
		return nil
	}

	tokens := tokenize(f.Expression)
	rules, err := parseTokens(tokens)
	if err != nil {
		return fmt.Errorf("failed to parse filter: %w", err)
	}

	f.rules = rules
	f.compiled = true
	return nil
}

// Match checks if a packet matches the filter
func (f *Filter) Match(packet *Packet) bool {
	if !f.compiled || len(f.rules) == 0 {
		return true // No filter = match all
	}

	for _, rule := range f.rules {
		if !f.matchRule(packet, &rule) {
			return false
		}
	}
	return true
}

func (f *Filter) matchRule(packet *Packet, rule *FilterRule) bool {
	result := false

	switch rule.Type {
	case FilterTypeProtocol:
		result = f.matchProtocol(packet, rule)
	case FilterTypeHost:
		result = f.matchHost(packet, rule)
	case FilterTypePort:
		result = f.matchPort(packet, rule)
	case FilterTypeNet:
		result = f.matchNet(packet, rule)
	case FilterTypePortRange:
		result = f.matchPortRange(packet, rule)
	}

	if rule.Negate {
		return !result
	}
	return result
}

func (f *Filter) matchProtocol(packet *Packet, rule *FilterRule) bool {
	switch strings.ToLower(rule.Protocol) {
	case "tcp":
		return packet.Type == PacketTypeTCP
	case "udp":
		return packet.Type == PacketTypeUDP
	case "icmp":
		return packet.Type == PacketTypeICMP
	case "arp":
		return packet.Type == PacketTypeARP
	}
	return false
}

func (f *Filter) matchHost(packet *Packet, rule *FilterRule) bool {
	if rule.SrcIP != nil {
		if rule.SrcIP.Contains(packet.SrcIP) || rule.SrcIP.Contains(packet.DstIP) {
			return true
		}
	}
	return false
}

func (f *Filter) matchPort(packet *Packet, rule *FilterRule) bool {
	switch rule.PortOp {
	case PortOpEqual:
		return packet.SrcPort == rule.SrcPort || packet.DstPort == rule.SrcPort
	case PortOpLess:
		return packet.SrcPort < rule.SrcPort || packet.DstPort < rule.SrcPort
	case PortOpGreater:
		return packet.SrcPort > rule.SrcPort || packet.DstPort > rule.SrcPort
	}
	return false
}

func (f *Filter) matchNet(packet *Packet, rule *FilterRule) bool {
	if rule.SrcIP != nil {
		return rule.SrcIP.Contains(packet.SrcIP) || rule.SrcIP.Contains(packet.DstIP)
	}
	return false
}

func (f *Filter) matchPortRange(packet *Packet, rule *FilterRule) bool {
	inRange := func(port uint16) bool {
		return port >= rule.SrcPort && port <= rule.DstPort
	}
	return inRange(packet.SrcPort) || inRange(packet.DstPort)
}

// tokenize splits the expression into tokens
func tokenize(expr string) []string {
	// Simple tokenizer for BPF-like expressions
	expr = strings.TrimSpace(expr)
	re := regexp.MustCompile(`\s+`)
	return re.Split(expr, -1)
}

// parseTokens parses tokens into filter rules
func parseTokens(tokens []string) ([]FilterRule, error) {
	var rules []FilterRule
	i := 0

	for i < len(tokens) {
		token := strings.ToLower(tokens[i])
		rule := FilterRule{}

		// Check for negation
		if token == "not" || token == "!" {
			rule.Negate = true
			i++
			if i >= len(tokens) {
				return nil, fmt.Errorf("unexpected end after 'not'")
			}
			token = strings.ToLower(tokens[i])
		}

		switch token {
		case "tcp", "udp", "icmp", "arp":
			rule.Type = FilterTypeProtocol
			rule.Protocol = token
			i++

		case "host":
			i++
			if i >= len(tokens) {
				return nil, fmt.Errorf("expected IP after 'host'")
			}
			ip := net.ParseIP(tokens[i])
			if ip == nil {
				return nil, fmt.Errorf("invalid IP: %s", tokens[i])
			}
			rule.Type = FilterTypeHost
			rule.SrcIP = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
			i++

		case "net":
			i++
			if i >= len(tokens) {
				return nil, fmt.Errorf("expected network after 'net'")
			}
			_, ipnet, err := net.ParseCIDR(tokens[i])
			if err != nil {
				return nil, fmt.Errorf("invalid network: %s", tokens[i])
			}
			rule.Type = FilterTypeNet
			rule.SrcIP = ipnet
			i++

		case "port":
			i++
			if i >= len(tokens) {
				return nil, fmt.Errorf("expected port after 'port'")
			}
			port, err := strconv.ParseUint(tokens[i], 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", tokens[i])
			}
			rule.Type = FilterTypePort
			rule.SrcPort = uint16(port)
			rule.PortOp = PortOpEqual
			i++

		case "portrange":
			i++
			if i >= len(tokens) {
				return nil, fmt.Errorf("expected port range after 'portrange'")
			}
			parts := strings.Split(tokens[i], "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", tokens[i])
			}
			start, err := strconv.ParseUint(parts[0], 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid start port: %s", parts[0])
			}
			end, err := strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid end port: %s", parts[1])
			}
			rule.Type = FilterTypePortRange
			rule.SrcPort = uint16(start)
			rule.DstPort = uint16(end)
			i++

		case "and", "&&":
			i++ // Skip logical operators (implicit AND)
			continue

		case "or", "||":
			i++ // OR not fully supported yet
			continue

		default:
			return nil, fmt.Errorf("unknown token: %s", token)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// String returns the filter expression
func (f *Filter) String() string {
	return f.Expression
}

// Validate validates the filter expression
func (f *Filter) Validate() error {
	if f.compiled {
		return nil
	}
	return f.compile()
}
