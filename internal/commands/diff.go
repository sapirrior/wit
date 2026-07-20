package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"wit/internal/compress"
	"wit/internal/diff"
	"wit/internal/xmlio"
)

type ArchiveItem struct {
	Name     string
	Path     string
	Mode     string
	Sha1     string
	Size     int64
	Target   string
	IsBinary bool
}

func getBinaryInfo(data []byte) string {
	if len(data) == 0 {
		return "empty"
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err == nil {
		return fmt.Sprintf("image/%s (%dx%d)", format, config.Width, config.Height)
	}
	return fmt.Sprintf("%d bytes", len(data))
}

func parseArchiveMetadataOnly(inXmlPath string) (map[string]ArchiveItem, error) {
	var tempFilePath string
	if strings.HasSuffix(inXmlPath, ".zip") {
		extracted, err := compress.UnzipFirst(inXmlPath, os.TempDir())
		if err != nil {
			return nil, fmt.Errorf("failed to extract zip: %w", err)
		}
		tempFilePath = extracted
		defer os.Remove(tempFilePath)
		inXmlPath = tempFilePath
	}

	xmlFile, err := os.Open(inXmlPath)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	items := make(map[string]ArchiveItem)

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}

		name := se.Name.Local
		if name != "file" && name != "binary" && name != "symlink" && name != "empty" {
			continue
		}

		var pathAttr, modeAttr, sha1Attr, sizeAttr, targetAttr string
		for _, attr := range se.Attr {
			val := xmlio.UnescapeAttr(attr.Value)
			switch attr.Name.Local {
			case "path":
				pathAttr = val
			case "mode":
				modeAttr = val
			case "sha1":
				sha1Attr = val
			case "size":
				sizeAttr = val
			case "target":
				targetAttr = val
			}
		}

		if pathAttr == "" {
			continue
		}

		item := ArchiveItem{
			Name:     name,
			Path:     pathAttr,
			Mode:     modeAttr,
			Sha1:     sha1Attr,
			Target:   targetAttr,
			IsBinary: (name == "binary"),
		}
		if sizeAttr != "" {
			item.Size, _ = strconv.ParseInt(sizeAttr, 10, 64)
		}

		if name == "file" || name == "binary" {
			// Skip CDATA content to keep memory usage minimal in metadata pass
			if err := decoder.Skip(); err != nil {
				return nil, err
			}
		}
		items[pathAttr] = item
	}

	return items, nil
}

func loadModifiedContents(inXmlPath string, modifiedPaths map[string]bool) (map[string][]byte, error) {
	var tempFilePath string
	if strings.HasSuffix(inXmlPath, ".zip") {
		extracted, err := compress.UnzipFirst(inXmlPath, os.TempDir())
		if err != nil {
			return nil, fmt.Errorf("failed to extract zip: %w", err)
		}
		tempFilePath = extracted
		defer os.Remove(tempFilePath)
		inXmlPath = tempFilePath
	}

	xmlFile, err := os.Open(inXmlPath)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	contents := make(map[string][]byte)

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}

		name := se.Name.Local
		if name != "file" && name != "binary" {
			continue
		}

		var pathAttr, encodingAttr string
		for _, attr := range se.Attr {
			val := xmlio.UnescapeAttr(attr.Value)
			switch attr.Name.Local {
			case "path":
				pathAttr = val
			case "encoding":
				encodingAttr = val
			}
		}

		if pathAttr == "" || !modifiedPaths[pathAttr] {
			if err := decoder.Skip(); err != nil {
				return nil, err
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
			return nil, innerErr
		}

		if encodingAttr == "base64" {
			cleanB64 := strings.Map(func(r rune) rune {
				if strings.ContainsRune(" \t\r\n", r) {
					return -1
				}
				return r
			}, string(cdataBuf))
			decoded, _ := base64.StdEncoding.DecodeString(cleanB64)
			contents[pathAttr] = decoded
		} else {
			contents[pathAttr] = cdataBuf
		}
	}

	return contents, nil
}

func Diff(archiveAPath, archiveBPath string) error {
	aItems, err := parseArchiveMetadataOnly(archiveAPath)
	if err != nil {
		return fmt.Errorf("error reading first archive: %w", err)
	}
	bItems, err := parseArchiveMetadataOnly(archiveBPath)
	if err != nil {
		return fmt.Errorf("error reading second archive: %w", err)
	}

	var added []string
	var removed []string
	var modified []string
	modifiedPathsMap := make(map[string]bool)

	for path := range bItems {
		if _, exists := aItems[path]; !exists {
			added = append(added, path)
		} else if aItems[path].Sha1 != bItems[path].Sha1 || aItems[path].Name != bItems[path].Name || aItems[path].Target != bItems[path].Target {
			modified = append(modified, path)
			modifiedPathsMap[path] = true
		}
	}

	for path := range aItems {
		if _, exists := bItems[path]; !exists {
			removed = append(removed, path)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(modified)

	// Stream only the modified contents into memory
	var aContents, bContents map[string][]byte
	if len(modifiedPathsMap) > 0 {
		aContents, err = loadModifiedContents(archiveAPath, modifiedPathsMap)
		if err != nil {
			return fmt.Errorf("error loading first archive modified contents: %w", err)
		}
		bContents, err = loadModifiedContents(archiveBPath, modifiedPathsMap)
		if err != nil {
			return fmt.Errorf("error loading second archive modified contents: %w", err)
		}
	}

	fmt.Printf("--- %s\n", archiveAPath)
	fmt.Printf("+++ %s\n\n", archiveBPath)

	if len(added) > 0 {
		fmt.Printf("Added (%d):\n", len(added))
		for _, p := range added {
			fmt.Printf("  + %s\n", p)
		}
		fmt.Println()
	}

	if len(removed) > 0 {
		fmt.Printf("Removed (%d):\n", len(removed))
		for _, p := range removed {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println()
	}

	if len(modified) > 0 {
		fmt.Printf("Modified (%d):\n", len(modified))
		for _, p := range modified {
			itemA := aItems[p]
			itemB := bItems[p]

			if itemA.IsBinary || itemB.IsBinary {
				infoA := getBinaryInfo(aContents[p])
				infoB := getBinaryInfo(bContents[p])
				fmt.Printf("Binary files %s and %s differ: %s -> %s\n", p, p, infoA, infoB)
				continue
			}

			if itemA.Name == "symlink" || itemB.Name == "symlink" {
				fmt.Printf("--- %s [symlink: %s]\n", p, itemA.Target)
				fmt.Printf("+++ %s [symlink: %s]\n", p, itemB.Target)
				continue
			}

			if itemA.Name == "empty" || itemB.Name == "empty" {
				fmt.Printf("--- %s [empty]\n", p)
				fmt.Printf("+++ %s\n", p)
				continue
			}

			linesA := strings.Split(string(aContents[p]), "\n")
			linesB := strings.Split(string(bContents[p]), "\n")

			if len(linesA) > 0 && linesA[len(linesA)-1] == "" {
				linesA = linesA[:len(linesA)-1]
			}
			if len(linesB) > 0 && linesB[len(linesB)-1] == "" {
				linesB = linesB[:len(linesB)-1]
			}

			edits := diff.Myers(linesA, linesB)
			diffOutput := diff.FormatUnified(p, edits)
			if diffOutput != "" {
				fmt.Print(diffOutput)
				fmt.Println()
			}
		}
	}

	return nil
}
