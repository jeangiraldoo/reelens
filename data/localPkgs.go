package data

import (
	"encoding/json"
)

// LocalPkgs holds every tracked package keyed by name.
type LocalPkgs map[string]LocalPkg

// Save writes the full package collection back to disk.
func (pkgs LocalPkgs) Save() error {
	data, err := json.MarshalIndent(pkgs, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(PkgDataFilePath, data)
}
