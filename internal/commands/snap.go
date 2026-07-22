package commands

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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

		// Handle directory-only excludes ending with /
		if strings.HasSuffix(patClean, "/") {
			dirPat := strings.TrimSuffix(patClean, "/")
			if strings.HasPrefix(relClean, dirPat+"/") || relClean == dirPat || strings.Contains(relClean, "/"+dirPat+"/") {
				return true
			}
		}

		if gitignore.MatchPattern(patClean, relClean) {
			return true
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

type processedResult struct {
	index        int
	err          error
	size         int64
	hash         string
	isBinary     bool
	data         []byte
	b64Data      string
	empty        bool
	modeStr      string
	escPath      string
	escTarget    string
	isSymlink    bool
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

	// Concurrent file processing
	numWorkers := 8
	if numWorkers > len(items) {
		numWorkers = len(items)
	}

	type workTask struct {
		index int
		item  PackedItem
	}

	tasks := make(chan workTask, len(items))
	results := make([]processedResult, len(items))
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				res := processedResult{
					index:   task.index,
					escPath: xmlio.EscapeAttr(task.item.path),
				}
				info, err := os.Lstat(task.item.absPath)
				if err != nil {
					res.err = err
					results[task.index] = res
					continue
				}

				res.modeStr = fmt.Sprintf("0%o", info.Mode().Perm())

				if task.item.isSymlink {
					res.isSymlink = true
					target, err := os.Readlink(task.item.absPath)
					if err != nil {
						target = ""
					}
					res.escTarget = xmlio.EscapeAttr(target)
				} else {
					size := info.Size()
					res.size = size
					if size == 0 {
						res.empty = true
					} else {
						hVal, err := fsutil.GetSHA1(task.item.absPath)
						if err != nil {
							res.err = err
							results[task.index] = res
							continue
						}
						res.hash = hVal

						binary := fsutil.IsBinary(task.item.absPath)
						res.isBinary = binary

						data, err := os.ReadFile(task.item.absPath)
						if err != nil {
							res.err = err
							results[task.index] = res
							continue
						}

						if binary {
							res.b64Data = base64.StdEncoding.EncodeToString(data)
						} else {
							res.data = data
						}
					}
				}
				results[task.index] = res
			}
		}()
	}

	for i, item := range items {
		tasks <- workTask{index: i, item: item}
	}
	close(tasks)
	wg.Wait()

	textCount, binaryCount, emptyCount, symlinkCount := 0, 0, 0, 0
	spinner := progress.NewSpinner()
	spinner.Start("Snapping")

	for i, item := range items {
		spinner.Update(i+1, len(items), item.path)
		res := results[i]
		if res.err != nil {
			fmt.Printf("\n  ! error processing: %s - %v\n", item.path, res.err)
			continue
		}

		if res.isSymlink {
			fmt.Fprintf(out, "  <symlink path=\"%s\" target=\"%s\" mode=\"%s\"/>\n", res.escPath, res.escTarget, res.modeStr)
			symlinkCount++
		} else if res.empty {
			fmt.Fprintf(out, "  <empty path=\"%s\" mode=\"%s\"/>\n", res.escPath, res.modeStr)
			emptyCount++
		} else if res.isBinary {
			fmt.Fprintf(out, "  <binary path=\"%s\" identity=\"%s\" mode=\"%s\" size=\"%d\" encoding=\"base64\">", res.escPath, res.hash, res.modeStr, res.size)
			xmlio.WriteCDATA(out, []byte(res.b64Data))
			fmt.Fprintf(out, "</binary>\n")
			binaryCount++
		} else {
			fmt.Fprintf(out, "  <file path=\"%s\" identity=\"%s\" mode=\"%s\" size=\"%d\">", res.escPath, res.hash, res.modeStr, res.size)
			xmlio.WriteCDATA(out, res.data)
			fmt.Fprintf(out, "</file>\n")
			textCount++
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

