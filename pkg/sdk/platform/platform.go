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

// Package platform provides cross-platform detection for OSAPI providers.
// The agent uses this package to select the correct provider variant
// based on OS family (debian, darwin, or generic linux).
package platform

import (
	"strings"

	"github.com/shirou/gopsutil/v4/host"
)

// HostInfoFn is the function used to retrieve host information.
// Override in tests to simulate different platforms.
var HostInfoFn = host.Info

// debianFamily lists distributions that belong to the Debian OS family
// and share the same provider implementations.
var debianFamily = map[string]bool{
	"ubuntu":   true,
	"debian":   true,
	"raspbian": true,
}

// Detect returns the OS family name for provider selection.
// Returns "debian", "darwin", or "" (generic linux/unknown).
func Detect() string {
	info, _ := HostInfoFn()
	if info == nil {
		return ""
	}

	platform := strings.ToLower(info.Platform)
	if platform == "" && strings.ToLower(info.OS) == "darwin" {
		return "darwin"
	}

	if debianFamily[platform] {
		return "debian"
	}

	return platform
}
