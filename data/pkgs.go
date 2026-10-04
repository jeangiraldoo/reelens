package data

import (
	"bytes"
	"encoding/json"
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

type LocalPkgs struct {
	Pkgs map[string]LocalPkg
}

func NewLocalPkgs() (LocalPkgs, error) {
	pkgs, err := readPkgs()
	if err != nil {
		return LocalPkgs{}, err
	}

	return LocalPkgs{
		Pkgs: pkgs,
	}, nil
}

// Write writes the full package collection back to disk.
func (localPkgs LocalPkgs) Write() error {
	data, err := json.MarshalIndent(localPkgs.Pkgs, "", "  ")
	if err != nil {
		return err
	}

	return mkFile(pkgsFileName, data)
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

// SetRelease replaces the in-memory entry for the fetched release, restamping FetchedAt.
// It does not persist; call Write on LocalPkgs to write the collection to disk.
func (localPkg *LocalPkg) SetRelease(release Release) {
	localPkg.FetchedAt = time.Now().Format(time.RFC3339)
	localPkg.Release = release
}

// ResolveVersionStatus compares the installed version against the latest known
// version and classifies the outcome. Non-semver values fall back to plain
// string equality, since their relative order cannot be determined.
func (localPkg LocalPkg) ResolveVersionStatus() ReleaseState {
	if localPkg.InstalledVersion == "" {
		return Unknown
	}

	latest, latestErr := semver.NewVersion(localPkg.LatestVersion)
	installed, installedErr := semver.NewVersion(localPkg.InstalledVersion)

	if latestErr != nil || installedErr != nil {
		if localPkg.InstalledVersion == localPkg.LatestVersion {
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

func readPkgs() (map[string]LocalPkg, error) {
	pkgsData := make(map[string]LocalPkg)

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

func getPkgsFilePath() (string, error) {
	dataDirPath, err := getDataDirPath()

	if err != nil {
		return "", err
	}

	path := filepath.Join(dataDirPath, pkgsFileName)
	return path, nil
}
