package github

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"reelens/providers"
	"strings"
)

type githubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
}

type githubTag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

type githubCommit struct {
	Commit struct {
		Message   string `json:"message"`
		Committer struct {
			Date string `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

// tagsPerPage is the page size requested from GitHub's tags endpoint; a
// shorter response marks the final page.
const tagsPerPage = 100

func getReleaseFromTag(repo string) (providers.Release, error) {
	var latestTag githubTag
	var latestVersion *semver.Version

	for page := 1; ; page++ {

		endpoint := fmt.Sprintf("tags?per_page=%d&page=%d", tagsPerPage, page)
		resp, err := getRequest(repo, endpoint)

		if err != nil {
			return providers.Release{}, err
		}

		var tags []githubTag

		err = json.NewDecoder(resp.Body).Decode(&tags)
		_ = resp.Body.Close()

		if err != nil {
			return providers.Release{}, err
		}

		// No more pages.
		if len(tags) == 0 {
			break
		}

		for _, tag := range tags {
			version, err := semver.NewVersion(tag.Name)
			if err != nil {
				continue
			}

			// Ignore prereleases such as alpha, beta and rc.
			if version.Prerelease() != "" {
				continue
			}

			if latestVersion == nil || version.GreaterThan(latestVersion) {
				latestVersion = version
				latestTag = tag
			}
		}

		// If fewer than a full page came back, this was the last one.
		if len(tags) < tagsPerPage {
			break
		}
	}

	if latestVersion == nil {
		return providers.Release{}, fmt.Errorf(
			"repository has no stable semantic version tags",
		)
	}

	resp, err := getRequest(repo, "commits/"+latestTag.Commit.SHA)

	if err != nil {
		return providers.Release{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	var commit githubCommit

	err = json.NewDecoder(resp.Body).Decode(&commit)
	if err != nil {
		return providers.Release{}, err
	}

	return providers.Release{
		Version:       latestTag.Name,
		Name:          strings.Split(commit.Commit.Message, "\n")[0],
		PublishedDate: commit.Commit.Committer.Date,
	}, nil
}

func getLatestFromRelease(repo string) (providers.Release, error) {
	resp, err := getRequest(repo, "releases/latest")
	if err != nil {
		return providers.Release{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	var release githubRelease

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return providers.Release{}, err
	}

	return providers.Release{
		Version:       release.TagName,
		Name:          release.Name,
		PublishedDate: release.PublishedAt,
	}, nil
}
