package commands

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"wit/internal/compress"
	"wit/internal/xmlio"
)

func Glance(inXmlPath, targetIdentity string) error {
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
	found := false

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

		var pathAttr, modeAttr, identityAttr, sizeAttr, encodingAttr, targetAttr string
		for _, attr := range se.Attr {
			val := xmlio.UnescapeAttr(attr.Value)
			switch attr.Name.Local {
			case "path":
				pathAttr = val
			case "mode":
				modeAttr = val
			case "identity":
				identityAttr = val
			case "size":
				sizeAttr = val
			case "encoding":
				encodingAttr = val
			case "target":
				targetAttr = val
			}
		}

		// Support prefix match (min 4 characters for safety/uniqueness)
		if len(targetIdentity) >= 4 && strings.HasPrefix(identityAttr, targetIdentity) {
			found = true
			fmt.Printf("File:     %s\n", pathAttr)
			fmt.Printf("Type:     %s\n", name)
			if sizeAttr != "" {
				fmt.Printf("Size:     %s bytes\n", sizeAttr)
			}
			fmt.Printf("Mode:     %s\n", modeAttr)
			fmt.Printf("Identity: %s\n", identityAttr)
			if name == "symlink" {
				fmt.Printf("Target:   %s\n", targetAttr)
			}

			if name == "file" || name == "binary" {
				var cdataBuf []byte
				var innerErr error
				for {
					innerT, err := decoder.Token()
					if err != nil {
						innerErr = err
						break
					}
					if ee, ok := innerT.(xml.EndElement); ok && ee.Name.Local == name {
						break
					}
					if cd, ok := innerT.(xml.CharData); ok {
						cdataBuf = append(cdataBuf, cd...)
					}
				}

				if innerErr != nil {
					return fmt.Errorf("error reading content: %w", innerErr)
				}

				var content []byte
				if encodingAttr == "base64" {
					cleanB64 := strings.Map(func(r rune) rune {
						if strings.ContainsRune(" \t\r\n", r) {
							return -1
						}
						return r
					}, string(cdataBuf))
					decoded, _ := base64.StdEncoding.DecodeString(cleanB64)
					content = decoded
				} else {
					content = cdataBuf
				}

				fmt.Println("\n--- Content ---")
				if encodingAttr == "base64" {
					fmt.Println("[base64 encoded binary data]")
					if len(cdataBuf) > 1000 {
						fmt.Println(string(cdataBuf[:1000]) + "\n... [truncated] ...")
					} else {
						fmt.Println(string(cdataBuf))
					}
				} else {
					fmt.Println(string(content))
				}
			}
			break
		} else {
			if name == "file" || name == "binary" {
				_ = decoder.Skip()
			}
		}
	}

	if !found {
		return fmt.Errorf("identity %q not found in archive", targetIdentity)
	}

	return nil
}
