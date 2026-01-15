package main

import (
	"flag"
	"fmt"
	"os"

	"portman/internal/rules"
)

func usage() {
	fmt.Print(`Usage:
  portman open   <port> <proto> --file <rules.v4>
  portman close  <port> <proto> --file <rules.v4>
  portman status <port> <proto> --file <rules.v4>

Proto:
  tcp | udp | tcp/udp

Examples:
  portman open 3306 tcp --file rules.v4
  portman close 3306 tcp --file rules.v4
  portman status 8000 tcp/udp --file rules.v4
`)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 4 {
		usage()
	}

	cmd := os.Args[1]
	portArg := os.Args[2]
	proto := os.Args[3]

	fs := flag.NewFlagSet("portman", flag.ExitOnError)
	file := fs.String("file", "", "path to rules.v4 file")

	_ = fs.Parse(os.Args[4:])

	if *file == "" {
		fmt.Println("Error: --file is required")
		usage()
	}

	var port int
	_, err := fmt.Sscanf(portArg, "%d", &port)
	if err != nil {
		fmt.Printf("Invalid port: %s\n", portArg)
		os.Exit(1)
	}

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
		if changed {
			if err := os.WriteFile(*file, []byte(out), 0644); err != nil {
				fmt.Printf("Failed to write file: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Rule opened.")
		} else {
			fmt.Println("Rule already open.")
		}

	case "close":
		out, changed, err := rules.Close(content, port, proto)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if changed {
			if err := os.WriteFile(*file, []byte(out), 0644); err != nil {
				fmt.Printf("Failed to write file: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Rule closed.")
		} else {
			fmt.Println("Rule already closed.")
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

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		usage()
	}
}
