package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reelens/utils/system"
)

const (
	dirName     = "reelens"
	pkgDataFile = "packages.json"
)

var PkgDataFilePath string

// Init resolves the state directory, creates it if needed, and sets the
// cache paths. It must be called before any command reads or writes the
// cache.
func Init() error {
	base, err := system.UserStateDir()

	if err != nil {
		return err
	}

	dataDirPath := filepath.Join(base, dirName)
	PkgDataFilePath = filepath.Join(dataDirPath, pkgDataFile)

	const cacheDirPerm = 0o755

	return os.MkdirAll(dataDirPath, cacheDirPerm)
}

func LoadPkgs() (LocalPkgs, error) {
	pkgsData := make(LocalPkgs)

	data, err := os.ReadFile(PkgDataFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return pkgsData, nil
		}
		return nil, fmt.Errorf("cannot read the release cache: %w", err)
	}

	// An empty file is equivalent to an empty cache.
	if len(bytes.TrimSpace(data)) == 0 {
		return pkgsData, nil
	}

	if err := json.Unmarshal(data, &pkgsData); err != nil {
		return nil, fmt.Errorf("cannot parse the release cache: %w", err)
	}

	for pkgName, pkgData := range pkgsData {
		pkgData.Name = pkgName
		pkgsData[pkgName] = pkgData
	}

	return pkgsData, nil
}

// Writes the file atomically: content is written to a temporary file beside
// the real one, then renamed over it.
func writeFile(filePath string, data []byte) error {
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
