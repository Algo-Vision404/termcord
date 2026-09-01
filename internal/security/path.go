package security

import (
	"os"
	"runtime"
)

const (
	DirPerm  = 0o700
	FilePerm = 0o600
)

// EnsureDir creates path with owner-only permissions on Unix.
func EnsureDir(path string) error {
	if path == "" || path == "." {
		return nil
	}
	if err := os.MkdirAll(path, DirPerm); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	return os.Chmod(path, DirPerm)
}

// SecureFile restricts a file to owner read/write on Unix.
func SecureFile(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	return os.Chmod(path, FilePerm)
}
