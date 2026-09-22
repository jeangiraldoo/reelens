package github

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reelens/config"
	"reelens/data"
	"reelens/providers"
)

type githubProvider struct{}

type githubConfig struct {
	RepoID string `yaml:"repoID"`
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

func (githubProvider) GetFile(repo string, fileName string) ([]byte, error) {
	branchName, err := getDefaultBranch(repo)

	if err != nil {
		return []byte{}, err
	}

	url := url.URL{Scheme: "https", Host: "raw.githubusercontent.com"}
	u := url.JoinPath(repo, branchName, fileName)

	res, err := http.Get(u.String())

	if err != nil {
		return []byte{}, err
	}

	defer res.Body.Close()
	return io.ReadAll(res.Body)
}
