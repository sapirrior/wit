package commands

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"wit/internal/fsutil"
	"wit/internal/gitignore"
	"wit/internal/progress"
	"wit/internal/xmlio"
)

type PackedItem struct {
	path      string
	absPath   string
	isSymlink bool
}

func matchesExclude(relPath string, excludes []string) bool {
	for _, pat := range excludes {
		patClean := filepath.ToSlash(pat)
		relClean := filepath.ToSlash(relPath)

		if strings.HasSuffix(patClean, "/") {
			dirPat := strings.TrimSuffix(patClean, "/")
			if strings.HasPrefix(relClean, dirPat+"/") || relClean == dirPat || strings.Contains(relClean, "/"+dirPat+"/") {
				return true
			}
		}

		if m, _ := filepath.Match(patClean, relClean); m {
			return true
		}

		parts := strings.Split(relClean, "/")
		for _, part := range parts {
			if m, _ := filepath.Match(patClean, part); m {
				return true
			}
		}
	}
	return false
}

func walkDirSorted(dirPath string, gi *gitignore.GitignoreContext, rootPath string, skipped *int, excludes []string) ([]PackedItem, error) {
	var items []PackedItem
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()
		if name == "." || name == ".." || name == ".git" {
			continue
		}

		fullPath := filepath.Join(dirPath, name)
		info, err := entry.Info()
		if err != nil {
			*skipped++
			continue
		}

		isSym := (info.Mode()&os.ModeSymlink != 0)
		isDir := entry.IsDir() && !isSym

		if gi.IsIgnored(fullPath, isDir) {
			*skipped++
			continue
		}

		rel, _ := filepath.Rel(rootPath, fullPath)
		relSlash := filepath.ToSlash(rel)
		if matchesExclude(relSlash, excludes) {
			*skipped++
			continue
		}

		if isDir {
			subItems, err := walkDirSorted(fullPath, gi, rootPath, skipped, excludes)
			if err == nil {
				items = append(items, subItems...)
			}
		} else {
			items = append(items, PackedItem{
				path:      relSlash,
				absPath:   fullPath,
				isSymlink: isSym,
			})
		}
	}
	return items, nil
}

func Snap(dirPath, outPath, message string, excludes []string, maxSize int64) error {
	absDir, err := filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	absDir = filepath.ToSlash(absDir)

	rootName := filepath.Base(absDir)
	if rootName == "" || rootName == "." || rootName == "/" {
		rootName = "wit"
	}

	fmt.Printf("Scanning  %s/\n", rootName)

	gi := gitignore.NewGitignoreContext(absDir)
	gi.LoadAll(absDir)

	skipped := 0
	items, err := walkDirSorted(absDir, gi, absDir, &skipped, excludes)
	if err != nil {
		return err
	}

	var totalBytes int64
	var filteredItems []PackedItem
	for _, item := range items {
		if !item.isSymlink {
			if info, err := os.Lstat(item.absPath); err == nil {
				if maxSize > 0 && info.Size() > maxSize {
					skipped++
					continue
				}
				totalBytes += info.Size()
			}
		}
		filteredItems = append(filteredItems, item)
	}
	items = filteredItems

	fmt.Printf("  %d files  (%.2f MB)  %d skipped\n", len(items), float64(totalBytes)/(1024.0*1024.0), skipped)

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	timeStr := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	fmt.Fprintf(out, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(out, "<wit version=\"3.0\" created=\"%s\" root=\"%s\" file_count=\"%d\" message=\"%s\">\n", timeStr, rootName, len(items), xmlio.EscapeAttr(message))

	textCount, binaryCount, emptyCount, symlinkCount := 0, 0, 0, 0

	spinner := progress.NewSpinner()
	spinner.Start("Snapping")

	for i, item := range items {
		spinner.Update(i+1, len(items), item.path)
		info, err := os.Lstat(item.absPath)
		if err != nil {
			fmt.Printf("\n  ! stat error: %s\n", item.path)
			continue
		}

		modeStr := fmt.Sprintf("0%o", info.Mode().Perm())
		escPath := xmlio.EscapeAttr(item.path)

		if item.isSymlink {
			target, err := os.Readlink(item.absPath)
			if err != nil {
				target = ""
			}
			escTarget := xmlio.EscapeAttr(target)
			fmt.Fprintf(out, "  <symlink path=\"%s\" target=\"%s\" mode=\"%s\"/>\n", escPath, escTarget, modeStr)
			symlinkCount++
		} else {
			size := info.Size()
			if size == 0 {
				fmt.Fprintf(out, "  <empty path=\"%s\" mode=\"%s\"/>\n", escPath, modeStr)
				emptyCount++
			} else {
				hVal, err := fsutil.GetSHA1(item.absPath)
				if err != nil {
					fmt.Printf("\n  ! hash error: %s\n", item.path)
					continue
				}

				binary := fsutil.IsBinary(item.absPath)
				data, err := os.ReadFile(item.absPath)
				if err != nil {
					fmt.Printf("\n  ! read error: %s\n", item.path)
					continue
				}

				if binary {
					b64 := base64.StdEncoding.EncodeToString(data)
					fmt.Fprintf(out, "  <binary path=\"%s\" sha1=\"%s\" mode=\"%s\" size=\"%d\" encoding=\"base64\">", escPath, hVal, modeStr, size)
					xmlio.WriteCDATA(out, []byte(b64))
					fmt.Fprintf(out, "</binary>\n")
					binaryCount++
				} else {
					fmt.Fprintf(out, "  <file path=\"%s\" sha1=\"%s\" mode=\"%s\" size=\"%d\">", escPath, hVal, modeStr, size)
					xmlio.WriteCDATA(out, data)
					fmt.Fprintf(out, "</file>\n")
					textCount++
				}
			}
		}
	}

	spinner.Stop()

	fmt.Fprintf(out, "</wit>\n")
	fmt.Printf("  text:%d  binary:%d  empty:%d  symlinks:%d\n", textCount, binaryCount, emptyCount, symlinkCount)

	outInfo, err := os.Stat(outPath)
	outSize := int64(0)
	if err == nil {
		outSize = outInfo.Size()
	}
	fmt.Printf("Done  ->  %s  (%.2f MB)\n", filepath.Base(outPath), float64(outSize)/(1024.0*1024.0))
	return nil
}
