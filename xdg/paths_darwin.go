package xdg

import (
	"os"
	"path/filepath"
)

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	return "/"
}

func DataDir() string {
	return xdgPath(envDataHome, homeAppSupport())
}

func ConfigDir() string {
	return xdgPath(envConfigHome, homeAppSupport())
}

func CacheDir() string {
	return xdgPath(envCacheHome, filepath.Join(home, "Library", "Caches"))
}

func homeAppSupport() string {
	return filepath.Join(home, "Library", "Application Support")
}
