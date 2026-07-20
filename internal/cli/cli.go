package cli

import "fmt"

const Version = "3.0.2"

func PrintVersion() {
	fmt.Printf("wit v%s\n", Version)
}

func PrintHelp() {
	fmt.Printf("wit v%s — AI-Native File Context Maker\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  wit <folder> [options]                     # Snap a folder")
	fmt.Println("  wit snap <folder> [options]                # Snap a folder (alias)")
	fmt.Println("  wit grab <file.xml> [-o <dir>]             # Restore a folder")
	fmt.Println("  wit verify <file.xml>                      # Verify integrity (hashes/sizes)")
	fmt.Println("  wit diff <a.xml> <b.xml>                   # Diff two snapshots")
	fmt.Println("  wit list <file.xml> [-l]                   # List file paths in archive")
	fmt.Println("  wit glance <file.xml> <identity>           # Inspect a specific file's content and metadata")
	fmt.Println("  wit patch <file.xml> [dir]                 # Apply only changed files")
	fmt.Println("  wit msg <file.xml>                         # View snap message")
	fmt.Println("  wit meta <file.xml>                        # View files & metadata")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -m <msg>            Add custom snapshot message")
	fmt.Println("  -o <path>           Custom destination file/directory (naming auto-appends .wit.xml)")
	fmt.Println("  -c                  Compress to .wit.xml.zip (snap only)")
	fmt.Println("  --exclude <pattern> Pattern to exclude files (e.g. *.log, node_modules/)")
	fmt.Println("  --max-size <size>   Skip files larger than limit (e.g. 1MB, 500KB, 100B)")
	fmt.Println("  -h, --help          Show this help menu")
	fmt.Println("  -v, --version       Show version info")
}
