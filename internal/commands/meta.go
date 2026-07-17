package commands

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"wit/internal/xmlio"
)

func Meta(inXmlPath string) error {
	xmlFile, err := os.Open(inXmlPath)
	if err != nil {
		return fmt.Errorf("fatal: %s not found", inXmlPath)
	}
	defer xmlFile.Close()

	fmt.Printf("Archive Metadata for: %s\n\n", filepath.Base(inXmlPath))

	decoder := xml.NewDecoder(xmlFile)
	var rootName, version, created, fileCount, message string
	witTagFound := false

	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		if se, ok := t.(xml.StartElement); ok && se.Name.Local == "wit" {
			witTagFound = true
			for _, attr := range se.Attr {
				switch attr.Name.Local {
				case "root":
					rootName = attr.Value
				case "version":
					version = attr.Value
				case "created":
					created = attr.Value
				case "file_count":
					fileCount = attr.Value
				case "message":
					message = attr.Value
				}
			}
			break
		}
	}

	if !witTagFound {
		return fmt.Errorf("fatal: Error: Invalid wit file format. Missing <wit> root element")
	}

	fmt.Printf("Root Directory: %s\n", rootName)
	fmt.Printf("Format Version: %s\n", version)
	fmt.Printf("Created At:     %s\n", created)
	fmt.Printf("File Count:     %s\n", fileCount)
	if message != "" {
		fmt.Printf("Message:        %s\n", message)
	}
	fmt.Println()
	fmt.Println("Files:")

	_, _ = xmlFile.Seek(0, 0)
	decoder = xml.NewDecoder(xmlFile)

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}

		name := se.Name.Local
		if name != "file" && name != "binary" && name != "symlink" && name != "empty" {
			continue
		}

		var pathAttr, modeAttr, sizeAttr, targetAttr string
		for _, attr := range se.Attr {
			val := xmlio.UnescapeAttr(attr.Value)
			switch attr.Name.Local {
			case "path":
				pathAttr = val
			case "mode":
				modeAttr = val
			case "size":
				sizeAttr = val
			case "target":
				targetAttr = val
			}
		}

		switch name {
		case "file":
			fmt.Printf("  [text]    %s  (size: %s, mode: %s)\n", pathAttr, sizeAttr, modeAttr)
		case "binary":
			fmt.Printf("  [binary]  %s  (size: %s, mode: %s)\n", pathAttr, sizeAttr, modeAttr)
		case "symlink":
			fmt.Printf("  [symlink] %s  -> %s  (mode: %s)\n", pathAttr, targetAttr, modeAttr)
		case "empty":
			fmt.Printf("  [empty]   %s  (mode: %s)\n", pathAttr, modeAttr)
		}
	}

	return nil
}
