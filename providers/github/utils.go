package github

import (
	"reelens/apiclient"
)

func (p githubProvider) getDefaultBranch(repo string) (string, error) {
	type info struct {
		DefaultBranch string `json:"default_branch"`
	}

	api := apiclient.New(dataAPIBase())

	i, err := api.GetJSON[info](repo, "")

	if err != nil {
		return "", err
	}

	return i.DefaultBranch, nil
}
