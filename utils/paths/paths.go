package pathUtils

import (
	"os"
	"path/filepath"
	"runtime"
)

// Base directory for app data that must survive restarts yet isn't worth
// backing up or syncing between hosts (XDG state semantics).
//
// Deliberately not a cache directory: cached releases keep the tool useful
// while offline, so this data must not live anywhere cleanup tools may
// consider purgeable.
func UserStateDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return dir, nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, "AppData", "Local"), nil

	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, "Library", "Application Support"), nil

	default:
		if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
			return dir, nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, ".local", "state"), nil
	}
}
