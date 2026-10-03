package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type checkStatus string

const (
	statusOK   checkStatus = "ok"
	statusWarn checkStatus = "warn"
	statusFail checkStatus = "fail"
)

type doctorCheck struct {
	Name    string      `json:"name"`
	Status  checkStatus `json:"status"`
	Message string      `json:"message"`
}

type doctorReport struct {
	Profile  string        `json:"profile"`
	Version  VersionInfo   `json:"version"`
	Checks   []doctorCheck `json:"checks"`
	Summary  checkSummary  `json:"summary"`
	Commands []commandRow  `json:"commands,omitempty"`
}

type checkSummary struct {
	OK    int `json:"ok"`
	Warn  int `json:"warn"`
	Fail  int `json:"fail"`
	Total int `json:"total"`
}

type commandRow struct {
	Name   string       `json:"name"`
	Class  commandClass `json:"class"`
	Reason string       `json:"reason,omitempty"`
}

var doctorShowCommands bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run environment and product-surface diagnostics",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
		report := runDoctorReport(doctorShowCommands)
		return writeDoctorReport(cmd.OutOrStdout(), report, jsonOutput)
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().BoolVar(&doctorShowCommands, "commands", false, "Include command stability inventory")
}

func runDoctorReport(includeCommands bool) doctorReport {
	report := doctorReport{
		Profile: runtimeProfile,
		Version: currentVersionInfo(),
		Checks: []doctorCheck{
			checkRuntimeProfile(),
			checkGoRuntime(),
			checkPathBinary("go", true),
			checkPathBinary("git", false),
			checkPathBinary("staticcheck", false),
			checkPathBinary("golangci-lint", false),
			checkHomeDirectory(),
			checkConfigFile(),
			checkLoopbackBind(),
		},
	}
	report.Summary = summarizeChecks(report.Checks)
	if includeCommands {
		report.Commands = commandRows()
	}
	return report
}

func writeDoctorReport(w interface{ Write([]byte) (int, error) }, report doctorReport, jsonOutput bool) error {
	if jsonOutput {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	fmt.Fprintf(w, "GoCat doctor (%s profile)\n\n", report.Profile)
	for _, check := range report.Checks {
		fmt.Fprintf(w, "[%s] %s: %s\n", strings.ToUpper(string(check.Status)), check.Name, check.Message)
	}
	fmt.Fprintf(w, "\nSummary: %d ok, %d warn, %d fail (%d total)\n",
		report.Summary.OK, report.Summary.Warn, report.Summary.Fail, report.Summary.Total)
	if len(report.Commands) > 0 {
		fmt.Fprintln(w, "\nCommand inventory:")
		for _, row := range report.Commands {
			if row.Reason == "" {
				fmt.Fprintf(w, "  %-14s %s\n", row.Name, row.Class)
			} else {
				fmt.Fprintf(w, "  %-14s %s - %s\n", row.Name, row.Class, row.Reason)
			}
		}
	}
	return nil
}

func checkRuntimeProfile() doctorCheck {
	if isValidProfile(runtimeProfile) {
		return doctorCheck{Name: "runtime profile", Status: statusOK, Message: runtimeProfile}
	}
	return doctorCheck{Name: "runtime profile", Status: statusFail, Message: "must be stable or experimental"}
}

func checkGoRuntime() doctorCheck {
	return doctorCheck{
		Name:    "go runtime",
		Status:  statusOK,
		Message: fmt.Sprintf("%s on %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
	}
}

func checkPathBinary(name string, required bool) doctorCheck {
	path, err := exec.LookPath(name)
	if err == nil {
		return doctorCheck{Name: name, Status: statusOK, Message: path}
	}
	if required {
		return doctorCheck{Name: name, Status: statusFail, Message: "not found in PATH"}
	}
	return doctorCheck{Name: name, Status: statusWarn, Message: "not found in PATH"}
}

func checkHomeDirectory() doctorCheck {
	home, err := os.UserHomeDir()
	if err != nil {
		return doctorCheck{Name: "home directory", Status: statusFail, Message: err.Error()}
	}
	if home == "" {
		return doctorCheck{Name: "home directory", Status: statusFail, Message: "empty home path"}
	}
	info, err := os.Stat(home)
	if err != nil {
		return doctorCheck{Name: "home directory", Status: statusFail, Message: err.Error()}
	}
	if !info.IsDir() {
		return doctorCheck{Name: "home directory", Status: statusFail, Message: "home path is not a directory"}
	}
	return doctorCheck{Name: "home directory", Status: statusOK, Message: home}
}

func checkConfigFile() doctorCheck {
	configPath, _ := rootCmd.PersistentFlags().GetString("config")
	if configPath == "" {
		return doctorCheck{Name: "config file", Status: statusOK, Message: "not configured"}
	}
	if _, err := os.Stat(configPath); err != nil {
		return doctorCheck{Name: "config file", Status: statusFail, Message: err.Error()}
	}
	if err := loadConfigFile(configPath); err != nil {
		return doctorCheck{Name: "config file", Status: statusFail, Message: err.Error()}
	}
	abs, _ := filepath.Abs(configPath)
	return doctorCheck{Name: "config file", Status: statusOK, Message: abs}
}

func checkLoopbackBind() doctorCheck {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return doctorCheck{Name: "loopback bind", Status: statusFail, Message: err.Error()}
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return doctorCheck{Name: "loopback bind", Status: statusOK, Message: addr}
}

func summarizeChecks(checks []doctorCheck) checkSummary {
	var summary checkSummary
	for _, check := range checks {
		switch check.Status {
		case statusOK:
			summary.OK++
		case statusWarn:
			summary.Warn++
		case statusFail:
			summary.Fail++
		}
		summary.Total++
	}
	return summary
}

func commandRows() []commandRow {
	rows := make([]commandRow, 0, len(commandInventory))
	for name, meta := range commandInventory {
		rows = append(rows, commandRow{Name: name, Class: meta.Class, Reason: meta.Reason})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows
}
