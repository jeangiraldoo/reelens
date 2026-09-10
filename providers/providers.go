package providers

import (
	"errors"
	"reelens/config"
	"reelens/data"
	"time"
)

type Provider interface {
	GetLatestRelease(pkgName string, pkg config.Package) (data.Release, error)
}

var CachePath string
var ReleaseCachePath string

var registry = map[string]Provider{}

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

	cache, err := data.LoadReleaseCache()
	if err != nil {
		return err
	}

	cache[packageName] = data.LocalPackageData{
		Release:          release,
		FetchedAt:        time.Now().Format(time.RFC3339),
		InstalledVersion: cache[packageName].InstalledVersion,
	}

	return data.SaveReleaseCache(cache)
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
