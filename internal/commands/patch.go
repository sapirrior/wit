package commands

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"wit/internal/compress"
	"wit/internal/fsutil"
	"wit/internal/xmlio"
)

func Patch(inXmlPath, destDirPath string) error {
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

	if destDirPath == "" {
		destDirPath = "."
	}

	absDestRoot, err := filepath.Abs(destDirPath)
	if err != nil {
		absDestRoot = destDirPath
	}
	absDestRoot = filepath.ToSlash(absDestRoot)

	fmt.Printf("Patching destination '%s' with archive '%s'\n\n", destDirPath, filepath.Base(inXmlPath))

	decoder := xml.NewDecoder(xmlFile)
	var rootDirName string
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
					rootDirName = attr.Value
				}
			}
			break
		}
	}

	if !witTagFound {
		return fmt.Errorf("fatal: Error: Invalid wit file format. Missing <wit> root element")
	}

	if destDirPath == "." && rootDirName != "" {
		// If destDirPath wasn't explicitly given as something else, we use the root directory name.
		// Wait, if destDirPath is ".", the user might want current dir or rootDirName folder.
		// Let's check if the directory rootDirName exists.
		if _, err := os.Stat(rootDirName); err == nil {
			absDestRoot, _ = filepath.Abs(rootDirName)
			absDestRoot = filepath.ToSlash(absDestRoot)
		}
	}

	_, _ = xmlFile.Seek(0, 0)
	decoder = xml.NewDecoder(xmlFile)

	createdCount, updatedCount, unchangedCount, errorCount := 0, 0, 0, 0

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

		var pathAttr, modeAttr, sha1Attr, encodingAttr, targetAttr string
		for _, attr := range se.Attr {
			val := xmlio.UnescapeAttr(attr.Value)
			switch attr.Name.Local {
			case "path":
				pathAttr = val
			case "mode":
				modeAttr = val
			case "sha1":
				sha1Attr = val
			case "encoding":
				encodingAttr = val
			case "target":
				targetAttr = val
			}
		}

		if pathAttr == "" || modeAttr == "" {
			fmt.Printf("  ! malformed tag: %s\n", pathAttr)
			errorCount++
			continue
		}

		if !isPathSafe(pathAttr) {
			fmt.Printf("  ! unsafe path: %s\n", pathAttr)
			errorCount++
			continue
		}

		fullDestPath := filepath.Join(absDestRoot, pathAttr)

		var modeOctal int64 = 0644
		if m, err := strconv.ParseInt(modeAttr, 8, 32); err == nil {
			modeOctal = m
		}

		// Check if file already exists
		info, statErr := os.Lstat(fullDestPath)
		exists := (statErr == nil)

		if name == "symlink" {
			needsCreation := true
			if exists && (info.Mode()&os.ModeSymlink != 0) {
				if t, err := os.Readlink(fullDestPath); err == nil && t == targetAttr {
					needsCreation = false
				}
			}
			if needsCreation {
				_ = os.Remove(fullDestPath)
				parentDir := filepath.Dir(fullDestPath)
				_ = os.MkdirAll(parentDir, 0777)
				if err := os.Symlink(targetAttr, fullDestPath); err == nil {
					if exists {
						fmt.Printf("  ~ updated  [symlink] %s -> %s\n", pathAttr, targetAttr)
						updatedCount++
					} else {
						fmt.Printf("  + created  [symlink] %s -> %s\n", pathAttr, targetAttr)
						createdCount++
					}
				} else {
					fmt.Printf("  ! failed   [symlink] %s\n", pathAttr)
					errorCount++
				}
			} else {
				fmt.Printf("  = unchanged [symlink] %s\n", pathAttr)
				unchangedCount++
			}
			continue
		}

		if name == "empty" {
			needsCreation := true
			if exists && info.Size() == 0 && !info.IsDir() {
				needsCreation = false
			}
			if needsCreation {
				_ = os.Remove(fullDestPath)
				parentDir := filepath.Dir(fullDestPath)
				_ = os.MkdirAll(parentDir, 0777)
				f, err := os.Create(fullDestPath)
				if err == nil {
					f.Close()
					_ = os.Chmod(fullDestPath, os.FileMode(modeOctal))
					if exists {
						fmt.Printf("  ~ updated  [empty]   %s\n", pathAttr)
						updatedCount++
					} else {
						fmt.Printf("  + created  [empty]   %s\n", pathAttr)
						createdCount++
					}
				} else {
					fmt.Printf("  ! failed   [empty]   %s\n", pathAttr)
					errorCount++
				}
			} else {
				fmt.Printf("  = unchanged [empty]   %s\n", pathAttr)
				unchangedCount++
			}
			continue
		}

		// Read inner CDATA content from the archive
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

		// Compare hash to see if unchanged
		if exists && !info.IsDir() {
			localHash, _ := fsutil.GetSHA1(fullDestPath)
			if localHash == sha1Attr {
				fmt.Printf("  = unchanged         %s\n", pathAttr)
				unchangedCount++
				continue
			}
		}

		// Overwrite or create file
		parentDir := filepath.Dir(fullDestPath)
		if err := os.MkdirAll(parentDir, 0777); err != nil {
			fmt.Printf("  ! mkdir error: %s\n", pathAttr)
			errorCount++
			continue
		}

		f, err := os.Create(fullDestPath)
		if err != nil {
			fmt.Printf("  ! write error: %s\n", pathAttr)
			errorCount++
			continue
		}
		_, err = f.Write(finalData)
		f.Close()
		_ = os.Chmod(fullDestPath, os.FileMode(modeOctal))

		if err == nil {
			if exists {
				fmt.Printf("  ~ updated           %s\n", pathAttr)
				updatedCount++
			} else {
				fmt.Printf("  + created           %s\n", pathAttr)
				createdCount++
			}
		} else {
			fmt.Printf("  ! failed write      %s\n", pathAttr)
			errorCount++
		}
	}

	fmt.Printf("\nPatch complete: %d created, %d updated, %d unchanged, %d errors\n", createdCount, updatedCount, unchangedCount, errorCount)
	if errorCount > 0 {
		return fmt.Errorf("errors occurred during patching")
	}
	return nil
}
