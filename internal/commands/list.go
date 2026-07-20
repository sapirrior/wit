package commands

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"wit/internal/compress"
	"wit/internal/xmlio"
)

func List(inXmlPath string, longFormat bool) error {
	var tempFilePath string
	if strings.HasSuffix(inXmlPath, ".zip") {
		extracted, err := compress.UnzipFirst(inXmlPath, os.TempDir())
		if err != nil {
			return fmt.Errorf("fatal: failed to extract zip archive: %s", err)
		}
		tempFilePath = extracted
		defer os.Remove(tempFilePath)
		inXmlPath = tempFilePath
	}

	xmlFile, err := os.Open(inXmlPath)
	if err != nil {
		return fmt.Errorf("fatal: %s not found", inXmlPath)
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)

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

		var pathAttr, modeAttr, sizeAttr, targetAttr, sha1Attr string
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
			case "sha1":
				sha1Attr = val
			}
		}

		if pathAttr == "" {
			continue
		}

		if longFormat {
			switch name {
			case "file":
				fmt.Printf("  [text]    %s  (size: %s, mode: %s, sha1: %s)\n", pathAttr, sizeAttr, modeAttr, sha1Attr)
			case "binary":
				fmt.Printf("  [binary]  %s  (size: %s, mode: %s, sha1: %s)\n", pathAttr, sizeAttr, modeAttr, sha1Attr)
			case "symlink":
				fmt.Printf("  [symlink] %s  -> %s  (mode: %s)\n", pathAttr, targetAttr, modeAttr)
			case "empty":
				fmt.Printf("  [empty]   %s  (mode: %s)\n", pathAttr, modeAttr)
			}
		} else {
			fmt.Println(pathAttr)
		}
	}

	return nil
}
