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
