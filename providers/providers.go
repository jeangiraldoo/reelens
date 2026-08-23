package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reelens/config"
	"reelens/utils/paths"
	"time"
)

type Release struct {
	Version       string `json:"version"`
	Name          string `json:"name"`
	PublishedDate string `json:"publishedDate"`
}

type ReleaseCache struct {
	Release

	CachedAt string `json:"cachedAt"`
}

type Provider interface {
	GetLatestRelease(pkgName string, pkg config.Package) (Release, error)
}

// Permissions for the cache directory and its files.
const (
	cacheDirPerm  = 0o755
	cacheFilePerm = 0o640
)

var CachePath string
var ReleaseCachePath string

var registry = map[string]Provider{}

// Init resolves the state directory, creates it if needed, and sets the
// cache paths. It must be called before any command reads or writes the
// cache.
func Init() error {
	base, err := pathUtils.UserStateDir()
	if err != nil {
		return err
	}

	CachePath = filepath.Join(base, "reelens")
	ReleaseCachePath = filepath.Join(CachePath, "release.json")

	return os.MkdirAll(CachePath, cacheDirPerm)
}

func Register(name string, p Provider) {
	registry[name] = p
}

func Lookup(name string) (Provider, error) {
	p, ok := registry[name]

	var err error
	if !ok {
		err = errors.New("unknown provider: " + name)
	}
	return p, err
}

func CacheRelease(packageName string, pkg config.Package) error {
	provider, err := Lookup(pkg.Provider.Type)
	if err != nil {
		return err
	}

	release, err := provider.GetLatestRelease(packageName, pkg)
	if err != nil {
		return err
	}

	cache, err := LoadReleaseCache()
	if err != nil {
		return err
	}

	cache[packageName] = ReleaseCache{
		Release:  release,
		CachedAt: time.Now().Format(time.RFC3339),
	}

	return saveReleaseCache(cache)
}

func LoadReleaseCache() (map[string]ReleaseCache, error) {
	cache := make(map[string]ReleaseCache)

	data, err := os.ReadFile(ReleaseCachePath)
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
func saveReleaseCache(cache map[string]ReleaseCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(ReleaseCachePath), "reelens-cache-*.json")
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

	if err := os.Rename(tmp.Name(), ReleaseCachePath); err != nil {
		return fmt.Errorf("cannot swap cache file into place: %w", err)
	}

	return nil
}

func DecodeProviderConfig[T any](pkgConfig config.Package) (cfg T, err error) {
	if pkgConfig.Provider.Node == nil {
		return cfg, errors.New("provider has no configuration")
	}

	if err = pkgConfig.Provider.Node.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
