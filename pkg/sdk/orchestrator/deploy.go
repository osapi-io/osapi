// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	defaultGitHubOwner = "retr0h"
	defaultGitHubRepo  = "osapi"
)

// releaseAsset represents a single asset in a GitHub release.
type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// githubRelease represents the GitHub API response for a release.
type githubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

// httpClient allows injection of a custom HTTP client for testing.
var httpClient = http.DefaultClient

// resolveLatestBinaryURL queries the GitHub API for the latest release
// of the osapi repository and returns the download URL for the binary
// matching the given OS and architecture.
func resolveLatestBinaryURL(
	ctx context.Context,
	goos string,
	goarch string,
) (string, error) {
	apiURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/latest",
		defaultGitHubOwner,
		defaultGitHubRepo,
	)

	return resolveFromURL(ctx, apiURL, goos, goarch)
}

// resolveFromURL fetches a GitHub release JSON from the given URL and
// returns the download URL for the binary matching goos/goarch.
func resolveFromURL(
	ctx context.Context,
	apiURL string,
	goos string,
	goarch string,
) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("query GitHub releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"GitHub releases returned %d (no release published?)",
			resp.StatusCode,
		)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode release response: %w", err)
	}

	return matchAsset(release.Assets, goos, goarch)
}

// matchAsset finds the download URL for the binary matching the given
// OS and architecture in the release assets.
func matchAsset(
	assets []releaseAsset,
	goos string,
	goarch string,
) (string, error) {
	suffix := fmt.Sprintf("_%s_%s", goos, goarch)

	for _, a := range assets {
		if strings.HasSuffix(a.Name, suffix) {
			return a.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf(
		"no osapi binary found for %s/%s in release assets",
		goos,
		goarch,
	)
}

// deployScript returns a shell script that ensures curl is available
// and downloads the osapi binary to /osapi inside the container.
func deployScript(
	binaryURL string,
) string {
	return fmt.Sprintf(
		`command -v curl >/dev/null 2>&1 || `+
			`(apt-get update -qq && apt-get install -y -qq ca-certificates curl >/dev/null 2>&1) && `+
			`curl -fsSL '%s' -o /osapi && chmod +x /osapi`,
		binaryURL,
	)
}
