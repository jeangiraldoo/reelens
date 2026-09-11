package providers

import (
	"errors"
	"reelens/config"
	"reelens/data"
	"time"
)

type Provider interface {
	GetLatestRelease(pkgName string, pkg config.Pkg) (data.Release, error)
}

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

func CacheRelease(pkgName string, pkg config.Pkg) error {
	provider, err := Lookup(pkg.Provider.Type)
	if err != nil {
		return err
	}

	release, err := provider.GetLatestRelease(pkgName, pkg)
	if err != nil {
		return err
	}

	pkgsData, err := data.LoadPkgs()
	if err != nil {
		return err
	}

	pkgsData[pkgName] = data.LocalPkg{
		Release:          release,
		FetchedAt:        time.Now().Format(time.RFC3339),
		InstalledVersion: pkgsData[pkgName].InstalledVersion,
	}

	return data.SaveReleaseCache(pkgsData)
}

func DecodeProviderConfig[T any](pkgConfig config.Pkg) (cfg T, err error) {
	if pkgConfig.Provider.Node == nil {
		return cfg, errors.New("provider has no configuration")
	}

	if err = pkgConfig.Provider.Node.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
