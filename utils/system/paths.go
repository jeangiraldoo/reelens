package system

import (
	"fmt"
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

// Writes the file atomically: content is written to a temporary file beside
// the real one, then renamed over it.
func WriteFile(filePath string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(filePath), "reelens-*.tmp")
	if err != nil {
		return fmt.Errorf("cannot create temporary file: %w", err)
	}
	// Best-effort cleanup on every exit path; after a successful rename the
	// old name no longer exists and this becomes a harmless no-op.
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("cannot write temporary file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot close temporary file: %w", err)
	}

	const tempPermissions = 0o640
	if err := os.Chmod(tmp.Name(), tempPermissions); err != nil {
		return fmt.Errorf("cannot set file permissions: %w", err)
	}
	if err := os.Rename(tmp.Name(), filePath); err != nil {
		return fmt.Errorf("cannot move file into place: %w", err)
	}

	return nil
}
