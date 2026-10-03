package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var interfacesCmd = &cobra.Command{
	Use:     "interfaces",
	Aliases: []string{"ifaces", "ifs", "ifconfig"},
	Short:   "List available network interfaces with IP addresses",
	Long: `Display all active network interfaces with their IPv4 addresses and MAC addresses.
Useful for determining which interface IP to use for reverse shell payloads.

Examples:
  gocat interfaces`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		PrintInterfaces()
	},
}

func init() {
	rootCmd.AddCommand(interfacesCmd)
}
