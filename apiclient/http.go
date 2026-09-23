package apiclient

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

type Client struct {
	url.URL
}

func New(base url.URL) *Client {
	return &Client{base}
}

// apiTimeout bounds every provider request. http.DefaultClient has none, so
// one stalled connection would otherwise hang every command forever; these
// calls normally return in well under a second.
const apiTimeout = 15 * time.Second

var httpClient = &http.Client{Timeout: apiTimeout}

// Get performs a GET against the client's base URL, joined with the repo and
// the requested endpoint. On success the response is handed to the caller,
// who must read and close its body. On any failure the body is drained and
// closed here and only an error comes back, so callers can safely ignore the
// response whenever err is non-nil.
func (client Client) Get(repo, endpoint string) (*http.Response, error) {
	endpointPath, query, _ := strings.Cut(endpoint, "?")
	u := client.JoinPath(repo, endpointPath)
	u.RawQuery = query

	resp, err := httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp, nil
	}

	const maxErrorBodyLen = 1024 // error payloads are tiny JSON snippets; never slurp more
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyLen))

	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		reset, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)

		mins := max(time.Until(time.Unix(reset, 0)), 0)

		return nil, fmt.Errorf(
			"rate limit exceeded, resets in %d min (or add authentication keys to your config)",
			mins,
		)
	}

	return nil, fmt.Errorf("unexpected status: %s: %s", resp.Status, bytes.TrimSpace(body))
}

func (client Client) GetJSON[T any](repo, endpoint string) (T, error) {
	resp, err := client.Get(repo, endpoint)
	if err != nil {
		var zero T
		return zero, err
	}
	defer resp.Body.Close()

	return decode[T](resp.Body)
}

func decode[T any](body io.Reader) (T, error) {
	var dat T
	if err := json.NewDecoder(body).Decode(&dat); err != nil {
		return dat, err
	}

	return dat, nil
}
