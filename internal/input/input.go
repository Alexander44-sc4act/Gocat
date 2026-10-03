package input

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// ParseHostPort parses host and port from command line arguments
// If only one argument is provided, it's treated as port with default host
func ParseHostPort(args []string, defaultHost string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("missing host and port")
	}

	if len(args) == 1 {
		// Only port provided
		port := args[0]
		if err := validatePort(port); err != nil {
			return "", "", err
		}
		return defaultHost, port, nil
	}

	if len(args) == 2 {
		// Host and port provided
		host, port := args[0], args[1]
		if err := validatePort(port); err != nil {
			return "", "", err
		}
		return host, port, nil
	}

	return "", "", fmt.Errorf("too many arguments")
}

// validatePort validates if the port is a valid number
func validatePort(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}

	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port number out of range (1-65535): %d", portNum)
	}

	return nil
}

// ParseCommand parses and validates command strings
func ParseCommand(command string) (string, []string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", nil
	}

	return parts[0], parts[1:]
}

// ValidateShell validates a shell program path or name.
//
// This is a UX pre-check, never a security boundary: every exec site must
// pass the shell to exec.Command directly (no intermediate shell), so
// metacharacters are rejected here to catch mistakes early rather than to
// contain an attacker. Bare names must resolve via PATH or belong to the
// known-shell set; anything else is rejected.
func ValidateShell(shell string) error {
	if shell == "" {
		return fmt.Errorf("shell cannot be empty")
	}

	for _, r := range shell {
		if unicode.IsControl(r) {
			return fmt.Errorf("shell path contains control characters")
		}
	}

	if strings.ContainsAny(shell, ";|&`$()<>\n\"'*?~#=") {
		return fmt.Errorf("shell path contains invalid characters")
	}

	if !strings.ContainsAny(shell, `/\`) {
		if _, err := exec.LookPath(shell); err == nil {
			return nil
		}
		switch strings.ToLower(filepath.Base(shell)) {
		case "sh", "bash", "dash", "zsh", "fish", "ksh", "cmd", "cmd.exe",
			"powershell", "powershell.exe", "pwsh", "pwsh.exe":
			return nil
		}
		return fmt.Errorf("shell %q not found in PATH", shell)
	}

	return nil
}
