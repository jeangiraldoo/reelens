package data

import (
	"errors"
	"github.com/Masterminds/semver/v3"
)

type LocalPkg struct {
	Release

	Name             string `json:"-"` // Set at runtime
	FetchedAt        string `json:"fetchedAt"`
	InstalledVersion string `json:"installedVersion"`
}

type Release struct {
	LatestVersion string `json:"latestVersion"`
	PublishedDate string `json:"publishedDate"`
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
