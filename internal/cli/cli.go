package cli

import "fmt"

const Version = "2.0.1"

func PrintVersion() {
	fmt.Printf("wit v%s\n", Version)
}

func PrintHelp() {
	fmt.Printf("wit v%s — AI-Native File Context Maker\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  wit <folder> [-m \"msg\"] [-o <file.xml>]   # Snap a folder")
	fmt.Println("  wit grab <file.xml> [-o <dir>]             # Restore a folder")
	fmt.Println("  wit msg <file.xml>                         # View snap message")
	fmt.Println("  wit meta <file.xml>                        # View files & metadata")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -m   Add custom snapshot message")
	fmt.Println("  -o   Custom destination file (snap) or directory (grab)")
	fmt.Println("  -h   Show this help menu")
	fmt.Println("  -v   Show version info")
}
