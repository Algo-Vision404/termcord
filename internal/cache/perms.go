package cache

import (
	"path/filepath"

	"github.com/termcord/termcord/internal/security"
)

func secureDir(path string) error {
	return security.EnsureDir(path)
}

func secureFile(path string) error {
	return security.SecureFile(path)
}

func parentDir(path string) string {
	return filepath.Dir(path)
}
