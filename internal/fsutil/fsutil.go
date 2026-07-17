package fsutil

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func GetSHA1(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func IsBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()

	buf := make([]byte, 8192)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return true
	}
	if n == 0 {
		return false
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}

	isUTF8 := true
	i := 0
	for i < n {
		if buf[i] <= 0x7F {
			i++
		} else if (buf[i] & 0xE0) == 0xC0 {
			if i+1 >= n || (buf[i+1]&0xC0) != 0x80 {
				isUTF8 = false
				break
			}
			i += 2
		} else if (buf[i] & 0xF0) == 0xE0 {
			if i+2 >= n || (buf[i+1]&0xC0) != 0x80 || (buf[i+2]&0xC0) != 0x80 {
				isUTF8 = false
				break
			}
			i += 3
		} else if (buf[i] & 0xF8) == 0xF0 {
			if i+3 >= n || (buf[i+1]&0xC0) != 0x80 || (buf[i+2]&0xC0) != 0x80 || (buf[i+3]&0xC0) != 0x80 {
				isUTF8 = false
				break
			}
			i += 4
		} else {
			isUTF8 = false
			break
		}
	}
	if isUTF8 {
		return false
	}

	highBytes := 0
	for j := 0; j < n; j++ {
		if buf[j] > 127 {
			highBytes++
		}
	}
	return (float64(highBytes) / float64(n)) > 0.30
}
