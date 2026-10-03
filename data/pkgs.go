package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"path/filepath"
	"reelens/utils/system"
	"time"
)

const pkgsFileName = "packages.json"

type Release struct {
	LatestVersion string `json:"latestVersion"`
	PublishedDate string `json:"publishedDate"`
}

func getPkgsFilePath() (string, error) {
	dataDirPath, err := getDataDirPath()

	if err != nil {
		return "", err
	}

	path := filepath.Join(dataDirPath, pkgsFileName)
	return path, nil
}

// LocalPkgs holds every tracked package keyed by name.
type LocalPkgs map[string]LocalPkg

func LoadPkgs() (LocalPkgs, error) {
	pkgsData := make(LocalPkgs)

	pkgsFilePath, err := getPkgsFilePath()
	if err != nil {
		return nil, err
	}

	data, err := system.ReadFile(pkgsFilePath)
	if err != nil {
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

// Save writes the full package collection back to disk.
func (pkgs LocalPkgs) Save() error {
	data, err := json.MarshalIndent(pkgs, "", "  ")
	if err != nil {
		return err
	}

	return mkFile(pkgsFileName, data)
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

// ReleaseState classifies where a package's installed version stands relative
// to its latest known version.
type ReleaseState int

const (
	Unknown ReleaseState = iota
	Current
	Outdated
	Ahead
)

type LocalPkg struct {
	Release

	Name             string `json:"-"` // Set at runtime
	FetchedAt        string `json:"fetchedAt"`
	InstalledVersion string `json:"installedVersion"`
}

// Save loads the full collection, replaces this entry, and writes it back.
// Prefer LocalPkgs.Save() when the caller already holds the loaded map.
func (pkg LocalPkg) Save() error {
	if pkg.Name == "" {
		return errors.New("package name not set")
	}

	pkgs, err := LoadPkgs()
	if err != nil {
		return err
	}

	pkgs[pkg.Name] = pkg

	return pkgs.Save()
}

// ResolveVersionStatus compares the installed version against the latest known
// version and classifies the outcome. Non-semver values fall back to plain
// string equality, since their relative order cannot be determined.
func (pkg LocalPkg) ResolveVersionStatus() ReleaseState {
	if pkg.InstalledVersion == "" {
		return Unknown
	}

	latest, latestErr := semver.NewVersion(pkg.LatestVersion)
	installed, installedErr := semver.NewVersion(pkg.InstalledVersion)

	if latestErr != nil || installedErr != nil {
		if pkg.InstalledVersion == pkg.LatestVersion {
			return Current
		}
		return Outdated
	}

	switch installed.Compare(latest) {
	case -1:
		return Outdated
	case 0:
		return Current
	default:
		return Ahead
	}
}
