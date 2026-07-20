package compress

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ZipFile creates a .zip archive at zipPath containing the file at srcPath.
func ZipFile(srcPath, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(srcPath)
	header.Method = zip.Deflate

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, srcFile)
	return err
}

// UnzipFirst extracts the first file in the zip archive to destDir and returns its absolute path.
func UnzipFirst(zipPath, destDir string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		return "", os.ErrNotExist
	}

	file := reader.File[0]
	rc, err := file.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	destPath := filepath.Join(destDir, filepath.Base(file.Name))
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, rc)
	if err != nil {
		return "", err
	}

	return destPath, nil
}
