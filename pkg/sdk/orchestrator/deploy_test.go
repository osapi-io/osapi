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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DeployTestSuite struct {
	suite.Suite
}

func TestDeployTestSuite(t *testing.T) {
	suite.Run(t, new(DeployTestSuite))
}

func (suite *DeployTestSuite) TestMatchAsset() {
	tests := []struct {
		name         string
		assets       []releaseAsset
		goos         string
		goarch       string
		validateFunc func(url string, err error)
	}{
		{
			name: "matches linux amd64 asset",
			assets: []releaseAsset{
				{Name: "osapi_1.0.0_darwin_all", BrowserDownloadURL: "https://example.com/darwin"},
				{
					Name:               "osapi_1.0.0_linux_amd64",
					BrowserDownloadURL: "https://example.com/linux_amd64",
				},
				{
					Name:               "osapi_1.0.0_linux_arm64",
					BrowserDownloadURL: "https://example.com/linux_arm64",
				},
			},
			goos:   "linux",
			goarch: "amd64",
			validateFunc: func(url string, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "https://example.com/linux_amd64", url)
			},
		},
		{
			name: "matches linux arm64 asset",
			assets: []releaseAsset{
				{
					Name:               "osapi_1.0.0_linux_amd64",
					BrowserDownloadURL: "https://example.com/linux_amd64",
				},
				{
					Name:               "osapi_1.0.0_linux_arm64",
					BrowserDownloadURL: "https://example.com/linux_arm64",
				},
			},
			goos:   "linux",
			goarch: "arm64",
			validateFunc: func(url string, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "https://example.com/linux_arm64", url)
			},
		},
		{
			name: "returns error when no matching asset",
			assets: []releaseAsset{
				{Name: "osapi_1.0.0_darwin_all", BrowserDownloadURL: "https://example.com/darwin"},
			},
			goos:   "linux",
			goarch: "amd64",
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "no osapi binary found for linux/amd64")
			},
		},
		{
			name:   "returns error when no assets",
			assets: nil,
			goos:   "linux",
			goarch: "amd64",
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "no osapi binary found")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			url, err := matchAsset(tc.assets, tc.goos, tc.goarch)
			tc.validateFunc(url, err)
		})
	}
}

func (suite *DeployTestSuite) TestResolveFromURL() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		goos         string
		goarch       string
		validateFunc func(url string, err error)
	}{
		{
			name: "resolves URL from release response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				release := githubRelease{
					TagName: "v1.0.0",
					Assets: []releaseAsset{
						{
							Name:               "osapi_1.0.0_linux_amd64",
							BrowserDownloadURL: "https://github.com/retr0h/osapi/releases/download/v1.0.0/osapi_1.0.0_linux_amd64",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(release)
			},
			goos:   "linux",
			goarch: "amd64",
			validateFunc: func(url string, err error) {
				assert.NoError(suite.T(), err)
				assert.Contains(suite.T(), url, "osapi_1.0.0_linux_amd64")
			},
		},
		{
			name: "returns error on 404",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			goos:   "linux",
			goarch: "amd64",
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "404")
			},
		},
		{
			name: "returns error on invalid JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{invalid`))
			},
			goos:   "linux",
			goarch: "amd64",
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "decode")
			},
		},
		{
			name: "returns error when no matching asset in release",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				release := githubRelease{
					TagName: "v1.0.0",
					Assets: []releaseAsset{
						{
							Name:               "osapi_1.0.0_darwin_all",
							BrowserDownloadURL: "https://example.com/darwin",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(release)
			},
			goos:   "linux",
			goarch: "amd64",
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "no osapi binary found")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			url, err := resolveFromURL(
				context.Background(),
				server.URL,
				tc.goos,
				tc.goarch,
			)
			tc.validateFunc(url, err)
		})
	}
}

func (suite *DeployTestSuite) TestResolveLatestBinaryURL() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		validateFunc func(url string, err error)
	}{
		{
			name: "resolves from GitHub API",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				release := githubRelease{
					TagName: "v1.0.0",
					Assets: []releaseAsset{
						{
							Name:               "osapi_1.0.0_linux_arm64",
							BrowserDownloadURL: "https://example.com/arm64",
						},
						{
							Name:               "osapi_1.0.0_linux_amd64",
							BrowserDownloadURL: "https://example.com/amd64",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(release)
			},
			validateFunc: func(url string, err error) {
				assert.NoError(suite.T(), err)
				assert.NotEmpty(suite.T(), url)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			original := httpClient
			// Create a client that redirects all requests to the test server.
			httpClient = &http.Client{
				Transport: &rewriteTransport{
					base: server.Client().Transport,
					url:  server.URL,
				},
			}
			defer func() { httpClient = original }()

			url, err := resolveLatestBinaryURL(
				context.Background(),
				"linux",
				"amd64",
			)
			tc.validateFunc(url, err)
		})
	}
}

// rewriteTransport redirects all requests to a test server URL.
type rewriteTransport struct {
	base http.RoundTripper
	url  string
}

func (t *rewriteTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = "http"
	req.URL.Host = t.url[len("http://"):]

	return t.base.RoundTrip(req)
}

func (suite *DeployTestSuite) TestResolveFromURLHTTPError() {
	tests := []struct {
		name         string
		apiURL       string
		validateFunc func(url string, err error)
	}{
		{
			name:   "returns error on connection failure",
			apiURL: "http://127.0.0.1:0/invalid",
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "query GitHub releases")
			},
		},
		{
			name:   "returns error on invalid URL",
			apiURL: "http://invalid\x00url",
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "create request")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			url, err := resolveFromURL(
				context.Background(),
				tc.apiURL,
				"linux",
				"amd64",
			)
			tc.validateFunc(url, err)
		})
	}
}

func (suite *DeployTestSuite) TestDeployScript() {
	tests := []struct {
		name         string
		binaryURL    string
		validateFunc func(script string)
	}{
		{
			name:      "generates valid download script",
			binaryURL: "https://example.com/osapi",
			validateFunc: func(script string) {
				assert.Contains(suite.T(), script, "curl")
				assert.Contains(suite.T(), script, "https://example.com/osapi")
				assert.Contains(suite.T(), script, "-o /osapi")
				assert.Contains(suite.T(), script, "chmod +x /osapi")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			script := deployScript(tc.binaryURL)
			tc.validateFunc(script)
		})
	}
}
