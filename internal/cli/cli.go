package cli

import "fmt"

const Version = "2.0.1"

func PrintVersion() {
	fmt.Printf("wit v%s\n", Version)
}

func PrintHelp() {
	fmt.Printf("wit v%s\n", Version)
	fmt.Println()
	fmt.Println("AI-Native high-fidelity XML snapshot system for directories.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("wit <folder> [-m \"message\"] [-o <output_file>]")
	fmt.Println("wit grab <archive> [-o <output_dir>]")
	fmt.Println("wit msg <archive>")
	fmt.Println("wit meta <archive>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("grab: Rebuild the folder structure from the XML archive")
	fmt.Println("msg: View the snapshot message saved inside the archive")
	fmt.Println("meta: View archive metadata and file list")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("-m: Add a custom snapshot message")
	fmt.Println("-o: Custom output destination file (for snap) or directory (for grab)")
	fmt.Println("-h, --help: Show this help message")
	fmt.Println("-v, --version: Show version info")
}
