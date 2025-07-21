package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// DirExists checks if a directory exists at the given path and is not a file.
func DirExists(path string) error {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", path)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}

// FileExists checks if a file exists at the given path and is not a directory.
func FileExists(path string) error {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", path)
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}

	return nil
}

// CopyFile copies a file from src to dst. If dst does not exist, it will be created.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close() //nolint:all

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close() //nolint:all

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return err
	}

	return nil
}

// CopyDir copies a directory and its contents from src to dst. If dst does not exist, it will be created.
func CopyDir(src, dst string) error {
	// Avoid copying hidden files and directories
	if strings.HasPrefix(path.Base(src), ".") {
		return nil
	}

	sourceDir, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceDir.Close() //nolint:all

	entries, err := sourceDir.Readdir(-1)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		srcPath := path.Join(src, entry.Name())
		dstPath := path.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CalculateChecksum calculates the SHA256 checksum of a file at the given path.
func CalculateChecksum(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:all

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, err
	}

	checksum := hasher.Sum(nil)
	return []byte(hex.EncodeToString(checksum)), nil
}

// CompareChecksum compares two byte slices for equality.
func CompareBytes(generated, provided []byte) bool {
	if len(generated) != len(provided) {
		return false
	}

	for i := range provided {
		if generated[i] != provided[i] {
			return false
		}
	}

	return true
}

// CompareChecksum compares the checksum of a file at the given path with the provided checksum.
func CompareFileChecksum(filePath string, checksum []byte) bool {
	fileChecksum, err := CalculateChecksum(filePath)
	if err != nil {
		return false
	}

	if len(fileChecksum) != len(checksum) {
		return false
	}

	for i := range checksum {
		if fileChecksum[i] != checksum[i] {
			return false
		}
	}

	return true
}
