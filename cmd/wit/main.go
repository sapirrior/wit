package main

import (
	"fmt"
	"os"
	"strings"
	"wit/internal/cli"
	"wit/internal/commands"
	"wit/internal/fsutil"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "fatal: Error: Target path required. Use 'wit -h' for usage.")
		os.Exit(1)
	}

	if args[1] == "-h" || args[1] == "--help" {
		cli.PrintHelp()
		os.Exit(0)
	}

	if args[1] == "-v" || args[1] == "--version" {
		cli.PrintVersion()
		os.Exit(0)
	}

	if args[1] == "snap" {
		var folder string
		var out string
		var message string
		for i := 2; i < len(args); i++ {
			if args[i] == "-o" {
				if i+1 < len(args) {
					out = args[i+1]
					i++
				} else {
					fmt.Fprintln(os.Stderr, "fatal: Error: -o option requires an argument.")
					os.Exit(1)
				}
			} else if args[i] == "-m" {
				if i+1 < len(args) {
					message = args[i+1]
					i++
				} else {
					fmt.Fprintln(os.Stderr, "fatal: Error: -m option requires an argument.")
					os.Exit(1)
				}
			} else if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "fatal: Error: Unknown option %s\n", args[i])
				os.Exit(1)
			} else {
				if folder != "" {
					fmt.Fprintf(os.Stderr, "fatal: Error: Multiple folders specified: %s and %s\n", folder, args[i])
					os.Exit(1)
				}
				folder = args[i]
			}
		}
		if folder == "" {
			fmt.Fprintln(os.Stderr, "fatal: Error: Target folder required for snap.")
			os.Exit(1)
		}
		if !fsutil.DirExists(folder) {
			fmt.Fprintf(os.Stderr, "fatal: Error: Target '%s' is not a directory.\n", folder)
			os.Exit(1)
		}
		if out == "" {
			out = "witFile.xml"
			if _, err := os.Stat(out); err == nil {
				n := 1
				for {
					candidate := fmt.Sprintf("witFile-%d.xml", n)
					if _, err := os.Stat(candidate); os.IsNotExist(err) {
						out = candidate
						break
					}
					n++
				}
			}
		}
		if err := commands.Snap(folder, out, message); err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if args[1] == "grab" {
		var archive string
		var outDir string
		for i := 2; i < len(args); i++ {
			if args[i] == "-o" {
				if i+1 < len(args) {
					outDir = args[i+1]
					i++
				} else {
					fmt.Fprintln(os.Stderr, "fatal: Error: -o option requires an argument.")
					os.Exit(1)
				}
			} else if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "fatal: Error: Unknown option %s\n", args[i])
				os.Exit(1)
			} else {
				if archive != "" {
					fmt.Fprintf(os.Stderr, "fatal: Error: Multiple archives specified: %s and %s\n", archive, args[i])
					os.Exit(1)
				}
				archive = args[i]
			}
		}
		if archive == "" {
			fmt.Fprintln(os.Stderr, "fatal: Error: Target archive required for grab.")
			os.Exit(1)
		}
		if err := commands.Grab(archive, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if args[1] == "msg" {
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "fatal: Error: Target archive required for msg.")
			os.Exit(1)
		}
		archive := args[2]
		if err := commands.Msg(archive); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if args[1] == "meta" {
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "fatal: Error: Target archive required for meta.")
			os.Exit(1)
		}
		archive := args[2]
		if err := commands.Meta(archive); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Fallback to snapping: wit <folder> [-m "message"] [-o <output_file>]
	var folder string
	var out string
	var message string

	for i := 1; i < len(args); i++ {
		if args[i] == "-o" {
			if i+1 < len(args) {
				out = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "fatal: Error: -o option requires an argument.")
				os.Exit(1)
			}
		} else if args[i] == "-m" {
			if i+1 < len(args) {
				message = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "fatal: Error: -m option requires an argument.")
				os.Exit(1)
			}
		} else if strings.HasPrefix(args[i], "-") {
			fmt.Fprintf(os.Stderr, "fatal: Error: Unknown option %s\n", args[i])
			os.Exit(1)
		} else {
			if folder != "" {
				fmt.Fprintf(os.Stderr, "fatal: Error: Multiple folders specified: %s and %s\n", folder, args[i])
				os.Exit(1)
			}
			folder = args[i]
		}
	}

	if folder == "" {
		fmt.Fprintln(os.Stderr, "fatal: Error: Target folder required.")
		os.Exit(1)
	}

	if fsutil.DirExists(folder) {
		if out == "" {
			out = "witFile.xml"
			if _, err := os.Stat(out); err == nil {
				n := 1
				for {
					candidate := fmt.Sprintf("witFile-%d.xml", n)
					if _, err := os.Stat(candidate); os.IsNotExist(err) {
						out = candidate
						break
					}
					n++
				}
			}
		}
		if err := commands.Snap(folder, out, message); err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %s\n", err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "fatal: Error: Target folder '%s' does not exist or is not a directory.\n", folder)
		os.Exit(1)
	}
}
