package data

import (
	"encoding/json"
	"time"

	"reelens/utils/system"
)

// LocalPkgs holds every tracked package keyed by name.
type LocalPkgs map[string]LocalPkg

// Save writes the full package collection back to disk.
func (pkgs LocalPkgs) Save() error {
	data, err := json.MarshalIndent(pkgs, "", "  ")
	if err != nil {
		return err
	}

	return system.WriteFile(PkgDataFilePath, data)
}

// SetRelease replaces the in-memory entry for pkgName with the fetched release,
// restamping FetchedAt and preserving the installed version. It does not persist;
// call Save to write the collection to disk.
func (pkgs LocalPkgs) SetRelease(pkgName string, release Release) {
	pkgs[pkgName] = LocalPkg{
		Name:             pkgName,
		Release:          release,
		FetchedAt:        time.Now().Format(time.RFC3339),
		InstalledVersion: pkgs[pkgName].InstalledVersion,
	}
}
