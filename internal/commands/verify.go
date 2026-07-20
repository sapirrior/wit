package commands

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"wit/internal/compress"
	"wit/internal/xmlio"
)

func Verify(inXmlPath string) error {
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

	fmt.Printf("Verifying  %s\n\n", filepath.Base(inXmlPath))

	decoder := xml.NewDecoder(xmlFile)
	var rootName string
	witTagFound := false

	// Scan first for <wit> element
	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		if se, ok := t.(xml.StartElement); ok && se.Name.Local == "wit" {
			witTagFound = true
			for _, attr := range se.Attr {
				if attr.Name.Local == "root" {
					rootName = attr.Value
				}
			}
			break
		}
	}

	if !witTagFound {
		return fmt.Errorf("fatal: Error: Invalid wit file format. Missing <wit> root element")
	}

	fmt.Printf("Root Directory: %s\n\n", rootName)

	_, _ = xmlFile.Seek(0, 0)
	decoder = xml.NewDecoder(xmlFile)

	verifiedCount, skippedCount, errorCount := 0, 0, 0

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorCount++
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

		var pathAttr, identityAttr, sizeAttr, encodingAttr string
		for _, attr := range se.Attr {
			val := xmlio.UnescapeAttr(attr.Value)
			switch attr.Name.Local {
			case "path":
				pathAttr = val
			case "identity":
				identityAttr = val
			case "size":
				sizeAttr = val
			case "encoding":
				encodingAttr = val
			}
		}

		if pathAttr == "" {
			fmt.Println("  ! malformed tag (missing path)")
			errorCount++
			continue
		}

		if name == "symlink" || name == "empty" {
			fmt.Printf("  ✓ %s (skipped content check)\n", pathAttr)
			skippedCount++
			continue
		}

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
			fmt.Printf("  ! Unexpected error reading data for %s\n", pathAttr)
			errorCount++
			continue
		}

		var finalData []byte
		if encodingAttr == "base64" {
			cleanB64 := strings.Map(func(r rune) rune {
				if strings.ContainsRune(" \t\r\n", r) {
					return -1
				}
				return r
			}, string(cdataBuf))
			decoded, err := base64.StdEncoding.DecodeString(cleanB64)
			if err != nil {
				fmt.Printf("  ! Failed to decode base64 for %s\n", pathAttr)
				errorCount++
				continue
			}
			finalData = decoded
		} else {
			finalData = cdataBuf
		}

		integrityOk := true
		if sizeAttr != "" {
			expectedSize, _ := strconv.ParseInt(sizeAttr, 10, 64)
			if expectedSize != int64(len(finalData)) {
				fmt.Printf("  ! size mismatch: %s\n", pathAttr)
				integrityOk = false
			}
		}

		if identityAttr != "" {
			h := sha1.New()
			h.Write(finalData)
			actualHash := hex.EncodeToString(h.Sum(nil))
			if identityAttr != actualHash {
				fmt.Printf("  ! hash mismatch: %s\n", pathAttr)
				integrityOk = false
			}
		}

		if integrityOk {
			fmt.Printf("  ✓ %s\n", pathAttr)
			verifiedCount++
		} else {
			errorCount++
		}
	}

	fmt.Printf("\nVerified %d files  %d skipped  %d errors\n", verifiedCount, skippedCount, errorCount)
	if errorCount > 0 {
		return fmt.Errorf("verification failed")
	}
	return nil
}
