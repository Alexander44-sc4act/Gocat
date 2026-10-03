package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type verifyReport struct {
	Profile string        `json:"profile"`
	Checks  []doctorCheck `json:"checks"`
	Summary checkSummary  `json:"summary"`
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run local self-tests for the GoCat CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
		report := runVerifyReport()
		if err := writeVerifyReport(cmd.OutOrStdout(), report, jsonOutput); err != nil {
			return err
		}
		if report.Summary.Fail > 0 {
			return fmt.Errorf("verify failed: %d check(s) failed", report.Summary.Fail)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

func runVerifyReport() verifyReport {
	checks := []doctorCheck{
		verifyProfile(),
		verifyCommandInventory(),
		verifyConfigIfProvided(),
		verifyTempWritable(),
		verifyLoopbackDial(),
	}
	return verifyReport{
		Profile: runtimeProfile,
		Checks:  checks,
		Summary: summarizeChecks(checks),
	}
}

func writeVerifyReport(w interface{ Write([]byte) (int, error) }, report verifyReport, jsonOutput bool) error {
	if jsonOutput {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}
	fmt.Fprintf(w, "GoCat verify (%s profile)\n\n", report.Profile)
	for _, check := range report.Checks {
		fmt.Fprintf(w, "[%s] %s: %s\n", strings.ToUpper(string(check.Status)), check.Name, check.Message)
	}
	fmt.Fprintf(w, "\nSummary: %d ok, %d warn, %d fail (%d total)\n",
		report.Summary.OK, report.Summary.Warn, report.Summary.Fail, report.Summary.Total)
	return nil
}

func verifyProfile() doctorCheck {
	if isValidProfile(runtimeProfile) {
		return doctorCheck{Name: "profile", Status: statusOK, Message: runtimeProfile}
	}
	return doctorCheck{Name: "profile", Status: statusFail, Message: "must be stable or experimental"}
}

func verifyCommandInventory() doctorCheck {
	registered := map[string]bool{}
	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden {
			continue
		}
		registered[cmd.Name()] = true
	}

	var missing []string
	for name := range commandInventory {
		if !registered[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return doctorCheck{
			Name:    "command inventory",
			Status:  statusFail,
			Message: "inventory names not registered: " + strings.Join(missing, ", "),
		}
	}
	return doctorCheck{Name: "command inventory", Status: statusOK, Message: fmt.Sprintf("%d commands classified", len(commandInventory))}
}

func verifyConfigIfProvided() doctorCheck {
	configPath, _ := rootCmd.PersistentFlags().GetString("config")
	if configPath == "" {
		return doctorCheck{Name: "config validation", Status: statusOK, Message: "not configured"}
	}
	if err := loadConfigFile(configPath); err != nil {
		return doctorCheck{Name: "config validation", Status: statusFail, Message: err.Error()}
	}
	abs, _ := filepath.Abs(configPath)
	return doctorCheck{Name: "config validation", Status: statusOK, Message: abs}
}

func verifyTempWritable() doctorCheck {
	f, err := os.CreateTemp("", "gocat-verify-*")
	if err != nil {
		return doctorCheck{Name: "temp writable", Status: statusFail, Message: err.Error()}
	}
	name := f.Name()
	if _, err := f.Write([]byte("ok")); err != nil {
		_ = f.Close()
		_ = os.Remove(name)
		return doctorCheck{Name: "temp writable", Status: statusFail, Message: err.Error()}
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return doctorCheck{Name: "temp writable", Status: statusFail, Message: err.Error()}
	}
	_ = os.Remove(name)
	return doctorCheck{Name: "temp writable", Status: statusOK, Message: os.TempDir()}
}

func verifyLoopbackDial() doctorCheck {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return doctorCheck{Name: "loopback dial", Status: statusFail, Message: err.Error()}
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		_ = conn.Close()
		done <- nil
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		return doctorCheck{Name: "loopback dial", Status: statusFail, Message: err.Error()}
	}
	_ = conn.Close()
	if err := <-done; err != nil {
		return doctorCheck{Name: "loopback dial", Status: statusFail, Message: err.Error()}
	}
	return doctorCheck{Name: "loopback dial", Status: statusOK, Message: ln.Addr().String()}
}
