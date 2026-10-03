//go:build !pcap
// +build !pcap

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var snifferCmd = &cobra.Command{
	Use:     "sniffer",
	Aliases: []string{"sniff", "capture"},
	Short:   "Network packet sniffer (requires pcap build tag)",
	Long: `Network packet sniffer is not available in this build.

To enable packet sniffing, rebuild with:
  make build-pcap
or
  go build -tags pcap

This feature requires libpcap to be installed on your system.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("❌ Sniffer feature not available in this build")
		fmt.Println("\nTo enable packet sniffing:")
		fmt.Println("  1. Install libpcap:")
		fmt.Println("     • macOS:   brew install libpcap")
		fmt.Println("     • Ubuntu:  sudo apt-get install libpcap-dev")
		fmt.Println("     • CentOS:  sudo yum install libpcap-devel")
		fmt.Println("\n  2. Rebuild with pcap support:")
		fmt.Println("     go build -tags pcap")
	},
}

func init() {
	rootCmd.AddCommand(snifferCmd)
}
