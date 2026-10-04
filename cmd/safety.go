package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type commandClass string

const (
	classStable       commandClass = "stable"
	classExperimental commandClass = "experimental"
	classDeprecated   commandClass = "deprecated"
)

type commandMeta struct {
	Class  commandClass `json:"class"`
	Reason string       `json:"reason,omitempty"`
}

var commandInventory = map[string]commandMeta{
	"benchmark":  {Class: classStable},
	"broker":     {Class: classStable},
	"chat":       {Class: classStable},
	"completion": {Class: classStable},
	"connect":    {Class: classStable},
	"convert":    {Class: classStable},
	"doctor":     {Class: classStable},
	"interfaces": {Class: classStable},
	"listen":     {Class: classStable},
	"metrics":    {Class: classStable},
	"proxy":      {Class: classStable},
	"scan":       {Class: classStable},
	"script":     {Class: classStable},
	"serve":      {Class: classStable},
	"sniffer":    {Class: classStable},
	"transfer":   {Class: classStable},
	"unix":       {Class: classStable},
	"verify":     {Class: classStable},
	"version":    {Class: classStable},
	"websocket":  {Class: classStable},

	"console":      {Class: classExperimental, Reason: "interactive session control is still being hardened"},
	"dns-tunnel":   {Class: classExperimental, Reason: "tunneling transports require explicit opt-in"},
	"mcp":          {Class: classExperimental, Reason: "MCP automation should be enabled deliberately"},
	"multi-listen": {Class: classExperimental, Reason: "multi-listener orchestration is experimental"},
	"payload":      {Class: classExperimental, Reason: "payload generation is outside the stable-safe profile"},
	"portforward":  {Class: classExperimental, Reason: "port forwarding is an explicit experimental capability"},
	"session":      {Class: classExperimental, Reason: "session management is still being hardened"},
	"stabilize":    {Class: classExperimental, Reason: "interactive shell stabilization requires explicit opt-in"},
	"tunnel":       {Class: classExperimental, Reason: "SSH tunneling is an explicit experimental capability"},
}

func isValidProfile(profile string) bool {
	switch profile {
	case profileStable, profileExperimental:
		return true
	default:
		return false
	}
}

func validateRuntimeProfile(cmd *cobra.Command, args []string) error {
	if !isValidProfile(runtimeProfile) {
		return fmt.Errorf("invalid profile %q (allowed: stable, experimental)", runtimeProfile)
	}
	return nil
}
