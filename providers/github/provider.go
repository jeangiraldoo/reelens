package github

import (
	"fmt"
	"reelens/config"
	"reelens/providers"
)

type githubProvider struct{}

type githubConfig struct {
	Type   string `yaml:"type"`
	RepoID string `yaml:"repoID"`
}

func init() {
	providers.Register("github", githubProvider{})
}

func (githubProvider) GetLatestRelease(pkgName string, pkgConfig config.Package) (release providers.Release, err error) {
	githubConfig, err := providers.DecodeProviderConfig[githubConfig](pkgConfig)
	if err != nil {
		return providers.Release{}, err
	}

	switch pkgConfig.Version {
	case "tag":
		release, err = getReleaseFromTag(githubConfig.RepoID)
	case "release":
		release, err = getLatestFromRelease(githubConfig.RepoID)
	default:
		release, err = providers.Release{}, fmt.Errorf("unknown version source: %s", pkgConfig.Version)
	}

	if err != nil {
		return providers.Release{}, err
	}

	// The timestamp travels exactly as GitHub sent it (RFC3339); formatting
	// happens at display time.
	return
}
