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
	"wit/internal/progress"
	"wit/internal/xmlio"
)

func isPathSafe(relPath string) bool {
	cleaned := filepath.Clean(relPath)
	return !strings.HasPrefix(cleaned, "..") && !filepath.IsAbs(cleaned)
}

func promptUser(action, path string) (rune, error) {
	fmt.Printf("%s '%s'? [y/n/a/q] (yes/no/all/quit): ", action, path)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		if err.Error() == "unexpected newline" {
			return 'n', nil
		}
		return 0, err
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if len(response) == 0 {
		return 'n', nil
	}
	return rune(response[0]), nil
}

func Grab(inXmlPath, outDirPath string, interactive bool) error {
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

	fmt.Printf("Rebuilding from  %s\n\n", filepath.Base(inXmlPath))

	decoder := xml.NewDecoder(xmlFile)
	var rootDirName string
	var destRoot string
	var fileCountVal string
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
				} else if attr.Name.Local == "file_count" {
					fileCountVal = attr.Value
				}
			}
			break
		}
	}

	if !witTagFound {
		return fmt.Errorf("fatal: Error: Invalid wit file format. Missing <wit> root element")
	}

	totalFiles := 0
	if fileCountVal != "" {
		totalFiles, _ = strconv.Atoi(fileCountVal)
	}

	if outDirPath != "" {
		destRoot = outDirPath
	} else if rootDirName != "" {
		destRoot = "./" + rootDirName
	} else {
		destRoot = "."
	}

	absDestRoot, err := filepath.Abs(destRoot)
	if err != nil {
		absDestRoot = destRoot
	}
	absDestRoot = filepath.ToSlash(absDestRoot)

	if entries, err := os.ReadDir(absDestRoot); err == nil && len(entries) > 0 && !interactive {
		fmt.Printf("  Destination '%s' is not empty. Overwriting files.\n", destRoot)
	}

	_, _ = xmlFile.Seek(0, 0)
	decoder = xml.NewDecoder(xmlFile)

	builtCount, skippedCount, errorCount := 0, 0, 0
	autoApply := !interactive

	var spinner *progress.Spinner
	if autoApply {
		spinner = progress.NewSpinner()
		spinner.Start("Rebuilding")
	}

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

		if spinner != nil {
			spinner.Update(builtCount+skippedCount+errorCount+1, totalFiles, pathAttr)
		}

		if pathAttr == "" || modeAttr == "" {
			fmt.Printf("\n  ! malformed tag: %s\n", pathAttr)
			errorCount++
			continue
		}

		if !isPathSafe(pathAttr) {
			fmt.Printf("\n  ! unsafe path: %s\n", pathAttr)
			skippedCount++
			continue
		}

		if !autoApply {
			ans, err := promptUser("Restore", pathAttr)
			if err != nil || ans == 'q' {
				fmt.Println("Aborted.")
				break
			}
			if ans == 'a' {
				autoApply = true
				spinner = progress.NewSpinner()
				spinner.Start("Rebuilding")
			} else if ans != 'y' {
				skippedCount++
				if name == "file" || name == "binary" {
					_ = decoder.Skip()
				}
				continue
			}
		}

		fullDestPath := filepath.Join(absDestRoot, pathAttr)
		parentDir := filepath.Dir(fullDestPath)
		if err := os.MkdirAll(parentDir, 0777); err != nil {
			fmt.Printf("\n  ! mkdir error: %s\n", pathAttr)
			errorCount++
			skippedCount++
			continue
		}

		var modeOctal int64 = 0644
		if m, err := strconv.ParseInt(modeAttr, 8, 32); err == nil {
			modeOctal = m
		}

		if name == "symlink" {
			_ = os.Remove(fullDestPath)
			if err := os.Symlink(targetAttr, fullDestPath); err == nil {
				builtCount++
			} else {
				fmt.Printf("\n  ! Cannot create symlink %s\n", pathAttr)
				errorCount++
			}
			continue
		}

		if name == "empty" {
			f, err := os.Create(fullDestPath)
			if err == nil {
				f.Close()
				_ = os.Chmod(fullDestPath, os.FileMode(modeOctal))
				builtCount++
			} else {
				fmt.Printf("\n  ! Failed to write %s\n", pathAttr)
				errorCount++
			}
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
			fmt.Printf("\n  ! Unexpected error reading data for %s\n", pathAttr)
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
				fmt.Printf("\n  ! Failed to decode base64 for %s\n", pathAttr)
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
				fmt.Printf("\n  ! size mismatch: %s\n", pathAttr)
				integrityOk = false
			}
		}

		if identityAttr != "" {
			h := sha1.New()
			h.Write(finalData)
			actualHash := hex.EncodeToString(h.Sum(nil))
			if identityAttr != actualHash {
				fmt.Printf("\n  ! hash mismatch: %s\n", pathAttr)
				integrityOk = false
			}
		}

		if !integrityOk {
			errorCount++
			continue
		}

		f, err := os.Create(fullDestPath)
		if err != nil {
			fmt.Printf("\n  ! write error: %s\n", pathAttr)
			errorCount++
			continue
		}
		_, err = f.Write(finalData)
		f.Close()
		_ = os.Chmod(fullDestPath, os.FileMode(modeOctal))

		if err == nil {
			builtCount++
		} else {
			fmt.Printf("\n  ! Failed to write %s\n", pathAttr)
			errorCount++
		}
	}

	if spinner != nil {
		spinner.Stop()
	}

	fmt.Printf("\nBuild complete  %d created  %d skipped  %d errors\n", builtCount, skippedCount, errorCount)
	if errorCount > 0 {
		return fmt.Errorf("errors occurred during rebuild")
	}
	return nil
}
