package github

import (
	"fmt"
	"io"
	"net/url"
	"reelens/apiclient"
	"reelens/config"
	"reelens/data"
	"reelens/providers"
)

type githubProvider struct{}

type githubConfig struct {
	RepoID string `yaml:"repoID"`
}

// dataAPIBase returns the GitHub API entry point; the repos/ path prefix is
// the provider-specific part of the URL scheme.
func dataAPIBase() url.URL {
	return url.URL{Scheme: "https", Host: "api.github.com", Path: "/repos"}
}

func rawAPIBase() url.URL {
	return url.URL{Scheme: "https", Host: "raw.githubusercontent.com"}
}

func init() {
	providers.Register("github", githubProvider{})
}

func (githubProvider) GetLatestRelease(pkgName string, pkgConfig config.Pkg) (release data.Release, err error) {
	githubConfig, err := providers.DecodeProviderConfig[githubConfig](pkgConfig)
	if err != nil {
		return data.Release{}, err
	}

	switch pkgConfig.Version {
	case "tag":
		release, err = getReleaseFromTag(githubConfig.RepoID)
	case "release":
		release, err = getLatestFromRelease(githubConfig.RepoID)
	default:
		release, err = data.Release{}, fmt.Errorf("unknown version source: %s", pkgConfig.Version)
	}

	if err != nil {
		return data.Release{}, err
	}

	// The timestamp travels exactly as GitHub sent it (RFC3339); formatting
	// happens at display time.
	return
}

func (p githubProvider) GetFile(repo string, fileName string) ([]byte, error) {
	branchName, err := p.getDefaultBranch(repo)

	if err != nil {
		return nil, err
	}

	api := apiclient.New(rawAPIBase())
	res, err := api.Get(repo, branchName+"/"+fileName)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return io.ReadAll(res.Body)
}
