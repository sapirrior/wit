package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"wit/internal/cli"
	"wit/internal/commands"
	"wit/internal/compress"
	"wit/internal/fsutil"
)

func enforceWitXmlExt(name string) string {
	if !strings.HasSuffix(name, ".wit.xml") {
		return name + ".wit.xml"
	}
	return name
}

func parseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	var multiplier int64 = 1
	var numStr string
	if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		numStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "B") {
		numStr = strings.TrimSuffix(sizeStr, "B")
	} else {
		numStr = sizeStr
	}
	val, err := strconv.ParseInt(strings.TrimSpace(numStr), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size limit: %s", sizeStr)
	}
	return val * multiplier, nil
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "error: target path required (use -h for help)")
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
		var excludes []string
		var maxSizeStr string
		var compressZip bool

		for i := 2; i < len(args); i++ {
			if args[i] == "-o" {
				if i+1 < len(args) {
					out = args[i+1]
					i++
				} else {
					fmt.Fprintln(os.Stderr, "error: -o requires an argument")
					os.Exit(1)
				}
			} else if args[i] == "-m" {
				if i+1 < len(args) {
					message = args[i+1]
					i++
				} else {
					fmt.Fprintln(os.Stderr, "error: -m requires an argument")
					os.Exit(1)
				}
			} else if args[i] == "--exclude" {
				if i+1 < len(args) {
					excludes = append(excludes, args[i+1])
					i++
				} else {
					fmt.Fprintln(os.Stderr, "error: --exclude requires an argument")
					os.Exit(1)
				}
			} else if args[i] == "--max-size" {
				if i+1 < len(args) {
					maxSizeStr = args[i+1]
					i++
				} else {
					fmt.Fprintln(os.Stderr, "error: --max-size requires an argument")
					os.Exit(1)
				}
			} else if args[i] == "-c" {
				compressZip = true
			} else if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "error: unknown option %s\n", args[i])
				os.Exit(1)
			} else {
				if folder != "" {
					fmt.Fprintf(os.Stderr, "error: multiple folders specified: %s and %s\n", folder, args[i])
					os.Exit(1)
				}
				folder = args[i]
			}
		}
		if folder == "" {
			fmt.Fprintln(os.Stderr, "error: target folder required")
			os.Exit(1)
		}
		if !fsutil.DirExists(folder) {
			fmt.Fprintf(os.Stderr, "error: '%s' is not a directory\n", folder)
			os.Exit(1)
		}

		if out != "" {
			out = enforceWitXmlExt(out)
		} else {
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

		var snapTarget string
		if compressZip {
			snapTarget = out + ".tmp"
		} else {
			snapTarget = out
		}

		maxSize, err := parseSize(maxSizeStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			os.Exit(1)
		}

		if err := commands.Snap(folder, snapTarget, message, excludes, maxSize); err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %s\n", err.Error())
			if compressZip {
				os.Remove(snapTarget)
			}
			os.Exit(1)
		}

		if compressZip {
			zipOut := out
			if !strings.HasSuffix(zipOut, ".zip") {
				zipOut = zipOut + ".zip"
			}
			if err := compress.ZipFile(snapTarget, zipOut); err != nil {
				fmt.Fprintf(os.Stderr, "fatal compression error: %s\n", err.Error())
				os.Remove(snapTarget)
				os.Exit(1)
			}
			os.Remove(snapTarget)
			outInfo, err := os.Stat(zipOut)
			outSize := int64(0)
			if err == nil {
				outSize = outInfo.Size()
			}
			fmt.Printf("Compressed ->  %s  (%.2f MB)\n", filepath.Base(zipOut), float64(outSize)/(1024.0*1024.0))
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
					fmt.Fprintln(os.Stderr, "error: -o requires an argument")
					os.Exit(1)
				}
			} else if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "error: unknown option %s\n", args[i])
				os.Exit(1)
			} else {
				if archive != "" {
					fmt.Fprintf(os.Stderr, "error: multiple archives specified: %s and %s\n", archive, args[i])
					os.Exit(1)
				}
				archive = args[i]
			}
		}
		if archive == "" {
			fmt.Fprintln(os.Stderr, "error: target archive required for grab")
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
			fmt.Fprintln(os.Stderr, "error: target archive required for msg")
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
			fmt.Fprintln(os.Stderr, "error: target archive required for meta")
			os.Exit(1)
		}
		archive := args[2]
		if err := commands.Meta(archive); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if args[1] == "verify" {
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "error: target archive required for verify")
			os.Exit(1)
		}
		archive := args[2]
		if err := commands.Verify(archive); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if args[1] == "diff" {
		if len(args) < 4 {
			fmt.Fprintln(os.Stderr, "error: two target archives required for diff")
			os.Exit(1)
		}
		archiveA := args[2]
		archiveB := args[3]
		if err := commands.Diff(archiveA, archiveB); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if args[1] == "list" {
		var archive string
		longFormat := false
		for i := 2; i < len(args); i++ {
			if args[i] == "-l" {
				longFormat = true
			} else {
				archive = args[i]
			}
		}
		if archive == "" {
			fmt.Fprintln(os.Stderr, "error: target archive required for list")
			os.Exit(1)
		}
		if err := commands.List(archive, longFormat); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if args[1] == "glance" {
		var archive string
		var identity string
		for i := 2; i < len(args); i++ {
			if args[i] == "-i" {
				if i+1 < len(args) {
					identity = args[i+1]
					i++
				} else {
					fmt.Fprintln(os.Stderr, "error: -i requires an argument")
					os.Exit(1)
				}
			} else {
				if archive == "" {
					archive = args[i]
				} else if identity == "" {
					identity = args[i]
				}
			}
		}
		if archive == "" || identity == "" {
			fmt.Fprintln(os.Stderr, "error: target archive and identity hash required for glance")
			os.Exit(1)
		}
		if err := commands.Glance(archive, identity); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if args[1] == "patch" {
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "error: target archive required for patch")
			os.Exit(1)
		}
		archive := args[2]
		destDir := "."
		if len(args) >= 4 {
			destDir = args[3]
		}
		if err := commands.Patch(archive, destDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Fallback to snapping: wit <folder> [options]
	var folder string
	var out string
	var message string
	var excludes []string
	var maxSizeStr string
	var compressZip bool

	for i := 1; i < len(args); i++ {
		if args[i] == "-o" {
			if i+1 < len(args) {
				out = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "error: -o requires an argument")
				os.Exit(1)
			}
		} else if args[i] == "-m" {
			if i+1 < len(args) {
				message = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "error: -m requires an argument")
				os.Exit(1)
			}
		} else if args[i] == "--exclude" {
			if i+1 < len(args) {
				excludes = append(excludes, args[i+1])
				i++
			} else {
				fmt.Fprintln(os.Stderr, "error: --exclude requires an argument")
				os.Exit(1)
			}
		} else if args[i] == "--max-size" {
			if i+1 < len(args) {
				maxSizeStr = args[i+1]
				i++
			} else {
				fmt.Fprintln(os.Stderr, "error: --max-size requires an argument")
				os.Exit(1)
			}
		} else if args[i] == "-c" {
			compressZip = true
		} else if strings.HasPrefix(args[i], "-") {
			fmt.Fprintf(os.Stderr, "error: unknown option %s\n", args[i])
			os.Exit(1)
		} else {
			if folder != "" {
				fmt.Fprintf(os.Stderr, "error: multiple folders specified: %s and %s\n", folder, args[i])
				os.Exit(1)
			}
			folder = args[i]
		}
	}

	if folder == "" {
		fmt.Fprintln(os.Stderr, "error: target folder required")
		os.Exit(1)
	}

	if fsutil.DirExists(folder) {
		if out != "" {
			out = enforceWitXmlExt(out)
		} else {
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

		var snapTarget string
		if compressZip {
			snapTarget = out + ".tmp"
		} else {
			snapTarget = out
		}

		maxSize, err := parseSize(maxSizeStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			os.Exit(1)
		}

		if err := commands.Snap(folder, snapTarget, message, excludes, maxSize); err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %s\n", err.Error())
			if compressZip {
				os.Remove(snapTarget)
			}
			os.Exit(1)
		}

		if compressZip {
			zipOut := out
			if !strings.HasSuffix(zipOut, ".zip") {
				zipOut = zipOut + ".zip"
			}
			if err := compress.ZipFile(snapTarget, zipOut); err != nil {
				fmt.Fprintf(os.Stderr, "fatal compression error: %s\n", err.Error())
				os.Remove(snapTarget)
				os.Exit(1)
			}
			os.Remove(snapTarget)
			outInfo, err := os.Stat(zipOut)
			outSize := int64(0)
			if err == nil {
				outSize = outInfo.Size()
			}
			fmt.Printf("Compressed ->  %s  (%.2f MB)\n", filepath.Base(zipOut), float64(outSize)/(1024.0*1024.0))
		}
	} else {
		fmt.Fprintf(os.Stderr, "error: '%s' does not exist or is not a directory\n", folder)
		os.Exit(1)
	}
}
