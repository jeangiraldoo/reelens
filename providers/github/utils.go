package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	host             = "api.github.com"
	maxErrorBodyLen  = 1024 // error payloads are tiny JSON snippets; never slurp more
	secondsPerMinute = 60
)

// apiTimeout bounds every GitHub request. http.DefaultClient has none, so
// one stalled connection would otherwise hang every command forever; these
// calls normally return in well under a second.
const apiTimeout = 15 * time.Second

var httpClient = &http.Client{Timeout: apiTimeout}

// Builds the API URL for a repository endpoint. Endpoints may carry a query
// string ("tags?per_page=100&page=1"), which is preserved verbatim; path
// segments are escaped properly instead of concatenated blind.
func apiURL(repo, endpoint string) string {
	base := url.URL{Scheme: "https", Host: host}

	path, query, _ := strings.Cut(endpoint, "?")
	u := base.JoinPath("repos", repo, path)
	u.RawQuery = query

	return u.String()
}

// Performs a GET against the GitHub API. On success the response is handed
// to the caller, who must read and close its body. On any failure the body
// is drained and closed here and only an error comes back, so callers can
// safely ignore the response whenever err is non-nil.
func getRequest(repo, endpoint string) (*http.Response, error) {
	resp, err := httpClient.Get(apiURL(repo, endpoint))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp, nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyLen))
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		reset, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)

		mins := (reset - time.Now().Unix()) / secondsPerMinute
		if mins < 0 {
			mins = 0
		}

		return nil, fmt.Errorf(
			"github: rate limit exceeded, resets in %d min (or add authentication keys to your config)",
			mins,
		)
	}

	return nil, fmt.Errorf("github: %s: %s", resp.Status, bytes.TrimSpace(body))
}

func getDefaultBranch(repo string) (string, error) {
	resp, err := getRequest(repo, "")
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var info struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	return info.DefaultBranch, nil
}
