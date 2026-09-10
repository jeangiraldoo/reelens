package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"os"
	"path/filepath"
	"reelens/utils/paths"
)

const (
	dirName          = "reelens"
	packagesFileName = "packages.json"
)

// ReleaseState classifies where a package's installed version stands relative
// to its latest known version.
type ReleaseState int

const (
	Unknown ReleaseState = iota
	Current
	Outdated
	Ahead
)

const (
	cacheFilePerm = 0o640
)

type Release struct {
	Version       string `json:"version"`
	Name          string `json:"name"`
	PublishedDate string `json:"publishedDate"`
}

type LocalPackageData struct {
	Release

	CachedAt         string `json:"cachedAt"`
	InstalledVersion string `json:"installedVersion"`
}

var PackageDataFilePath string

// Init resolves the state directory, creates it if needed, and sets the
// cache paths. It must be called before any command reads or writes the
// cache.
func Init() error {
	base, err := pathUtils.UserStateDir()

	if err != nil {
		return err
	}

	dataDirPath := filepath.Join(base, dirName)
	PackageDataFilePath = filepath.Join(dataDirPath, packagesFileName)

	const cacheDirPerm = 0o755

	return os.MkdirAll(dataDirPath, cacheDirPerm)
}

// ResolveVersionStatus compares the installed version against the latest known
// version and classifies the outcome. Non-semver values fall back to plain
// string equality, since their relative order cannot be determined.
func (r LocalPackageData) ResolveVersionStatus() ReleaseState {
	if r.InstalledVersion == "" {
		return Unknown
	}

	latest, latestErr := semver.NewVersion(r.Version)
	installed, installedErr := semver.NewVersion(r.InstalledVersion)

	if latestErr != nil || installedErr != nil {
		if r.InstalledVersion == r.Version {
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

func SetCachedReleaseInstalledVersion(pkgName, newVersion string) error {
	cache, err := LoadReleaseCache()

	if err != nil {
		return err
	}

	cachePkgData, ok := cache[pkgName]

	if !ok {
		return errors.New("unknown package: " + pkgName)
	}

	cachePkgData.InstalledVersion = newVersion

	cache[pkgName] = cachePkgData

	return SaveReleaseCache(cache)
}

func LoadReleaseCache() (map[string]LocalPackageData, error) {
	cache := make(map[string]LocalPackageData)

	data, err := os.ReadFile(PackageDataFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, fmt.Errorf("cannot read the release cache: %w", err)
	}

	// An empty file is equivalent to an empty cache.
	if len(bytes.TrimSpace(data)) == 0 {
		return cache, nil
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("cannot parse the release cache: %w", err)
	}

	return cache, nil
}

// Persists the cache atomically: content is written to a temporary file
// beside the real one, then renamed over it. Rename is only atomic within
// one filesystem — hence the sibling placement — so readers always see
// either the complete old file or the complete new one, never a torn mix.
func SaveReleaseCache(cache map[string]LocalPackageData) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(PackageDataFilePath), "reelens-cache-*.json")
	if err != nil {
		return fmt.Errorf("cannot create temporary cache file: %w", err)
	}
	// Best-effort cleanup on every exit path; after a successful rename the
	// old name no longer exists and this becomes a harmless no-op.
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("cannot write temporary cache file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot close temporary cache file: %w", err)
	}

	if err := os.Chmod(tmp.Name(), cacheFilePerm); err != nil {
		return fmt.Errorf("cannot set cache file permissions: %w", err)
	}

	if err := os.Rename(tmp.Name(), PackageDataFilePath); err != nil {
		return fmt.Errorf("cannot swap cache file into place: %w", err)
	}

	return nil
}
