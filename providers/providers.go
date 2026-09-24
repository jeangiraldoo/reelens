package providers

import (
	"errors"
	"reelens/config"
	"reelens/data"
)

type Provider interface {
	GetLatestRelease(pkgName string, pkg config.Pkg) (data.Release, error)
	GetFile(repoID string, fileName string) ([]byte, error)
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

func DecodeProviderConfig[T any](pkgConfig config.Pkg) (cfg T, err error) {
	if pkgConfig.Provider.Node == nil {
		return cfg, errors.New("provider has no configuration")
	}

	if err = pkgConfig.Provider.Node.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
