package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"portman/internal/rules"
)

func usage() {
	fmt.Print(`Usage:
  portman open   <port> <proto> [--file <rules.v4>] [--dry-run] [--apply]
  portman close  <port> <proto> [--file <rules.v4>] [--dry-run] [--apply]
  portman status <port> <proto> [--file <rules.v4>]
	portman list                [--file <rules.v4>]

Proto:
  tcp | udp | tcp/udp

Flags:
  --file     Path to rules.v4 (default: /etc/iptables/rules.v4 on Linux)
  --dry-run  Do not write the file
  --apply    (Linux) Validate and apply using iptables-restore

Examples:
  portman open 3306 tcp
  portman open 8000 tcp/udp --apply
  portman close 3306 tcp --apply
  portman status 443 tcp
	portman list
`)
	os.Exit(1)
}

func defaultRulesPath() string {
	if runtime.GOOS == "linux" {
		return "/etc/iptables/rules.v4"
	}
	return ""
}

func isRootLinux() bool {
	if runtime.GOOS != "linux" {
		return true
	}
	return os.Geteuid() == 0
}

func backupPath(original string) string {
	dir := filepath.Dir(original)
	base := filepath.Base(original)
	ts := time.Now().Format("20060102-150405")
	return filepath.Join(dir, fmt.Sprintf("%s.bak-%s", base, ts))
}

func iptablesRestoreTest(file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	cmd := exec.Command("iptables-restore", "-t")
	cmd.Stdin = bytes.NewReader(b)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("iptables-restore -t failed: %v\n%s", err, string(out))
	}
	return nil
}

func iptablesRestoreApply(file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	cmd := exec.Command("iptables-restore")
	cmd.Stdin = bytes.NewReader(b)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("iptables-restore failed: %v\n%s", err, string(out))
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	cmd := os.Args[1]

	switch cmd {
	case "list":
		fs := flag.NewFlagSet("portman list", flag.ExitOnError)
		file := fs.String("file", defaultRulesPath(), "path to rules.v4")
		_ = fs.Parse(os.Args[2:])

		if *file == "" {
			fmt.Println("Error: --file is required on non-Linux systems")
			os.Exit(1)
		}

		contentBytes, err := os.ReadFile(*file)
		if err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			os.Exit(1)
		}

		items, err := rules.List(string(contentBytes))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if len(items) == 0 {
			fmt.Println("No open ports in PORTMAN block.")
			return
		}
		for _, it := range items {
			fmt.Printf("%d/%s\n", it.Port, it.Proto)
		}
		return
	case "open", "close", "status":
		// handled below
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
	}

	if len(os.Args) < 4 {
		usage()
	}

	portArg := os.Args[2]
	proto := os.Args[3]

	fs := flag.NewFlagSet("portman", flag.ExitOnError)
	file := fs.String("file", defaultRulesPath(), "path to rules.v4")
	dryRun := fs.Bool("dry-run", false, "do not write file")
	apply := fs.Bool("apply", false, "validate and apply with iptables-restore (Linux)")
	_ = fs.Parse(os.Args[4:])

	if *file == "" {
		fmt.Println("Error: --file is required on non-Linux systems")
		os.Exit(1)
	}

	if *apply && runtime.GOOS != "linux" {
		fmt.Println("Error: --apply is only supported on Linux")
		os.Exit(1)
	}

	if runtime.GOOS == "linux" && (cmd == "open" || cmd == "close") && (*apply || strings.HasPrefix(*file, "/etc/")) && !isRootLinux() {
		fmt.Println("Error: run as root (sudo) to modify rules and apply iptables-restore")
		os.Exit(1)
	}

	// Parse port
	var port int
	_, err := fmt.Sscanf(portArg, "%d", &port)
	if err != nil {
		fmt.Printf("Invalid port: %s\n", portArg)
		os.Exit(1)
	}

	// Read file
	contentBytes, err := os.ReadFile(*file)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
	}
	content := string(contentBytes)

	switch cmd {
	case "open":
		out, changed, err := rules.Open(content, port, proto)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if !changed {
			fmt.Println("Rule already open.")
			return
		}

		if *dryRun {
			fmt.Println("Dry run: changes would be written.")
			return
		}

		bak := backupPath(*file)
		if err := os.WriteFile(bak, contentBytes, 0644); err != nil {
			fmt.Printf("Failed to write backup: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(*file, []byte(out), 0644); err != nil {
			fmt.Printf("Failed to write file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Rule opened.\n")

		if *apply {
			if err := iptablesRestoreTest(*file); err != nil {
				fmt.Printf("Apply aborted (validation failed): %v\n", err)
				os.Exit(1)
			}
			if err := iptablesRestoreApply(*file); err != nil {
				fmt.Printf("Apply failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Applied with iptables-restore.")
		}

	case "close":
		out, changed, err := rules.Close(content, port, proto)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if !changed {
			fmt.Println("Rule already closed (or not managed by PORTMAN block).")
			return
		}

		if *dryRun {
			fmt.Println("Dry run: changes would be written.")
			return
		}

		bak := backupPath(*file)
		if err := os.WriteFile(bak, contentBytes, 0644); err != nil {
			fmt.Printf("Failed to write backup: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(*file, []byte(out), 0644); err != nil {
			fmt.Printf("Failed to write file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Rule closed.\n")

		if *apply {
			if err := iptablesRestoreTest(*file); err != nil {
				fmt.Printf("Apply aborted (validation failed): %v\n", err)
				os.Exit(1)
			}
			if err := iptablesRestoreApply(*file); err != nil {
				fmt.Printf("Apply failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Applied with iptables-restore.")
		}

	case "status":
		st, err := rules.Status(content, port, proto)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		for p, ok := range st {
			if ok {
				fmt.Printf("%s: open\n", p)
			} else {
				fmt.Printf("%s: closed\n", p)
			}
		}
	// same as just "portman"
	case "help":
		usage()

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
	}
}
