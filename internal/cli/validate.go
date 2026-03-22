// Copyright (c) 2024 John Dewey

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

package cli

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v4/host"
)

// hostInfoFn allows tests to override host.Info.
var hostInfoFn = host.Info

// OSFamily represents a supported OS family with its member distributions.
type OSFamily struct {
	// Name is the family name following Ansible conventions (e.g., "Debian", "RedHat").
	Name string
	// Distributions maps distribution names to their supported versions.
	Distributions map[string][]string
}

// supportedFamilies defines the OS families and distributions OSAPI supports.
var supportedFamilies = []OSFamily{
	{
		Name: "Debian",
		Distributions: map[string][]string{
			"debian": {"12", "13"},
			"ubuntu": {"20.04", "22.04", "24.04"},
		},
	},
}

// IsOSFamilySupported checks if the given distribution and version belong
// to a supported OS family. Returns the family name and true if supported.
func IsOSFamilySupported(
	distro string,
	version string,
) (string, bool) {
	distro = strings.ToLower(distro)

	for _, family := range supportedFamilies {
		versions, ok := family.Distributions[distro]
		if !ok {
			continue
		}

		for _, v := range versions {
			if v == version || strings.HasPrefix(version, v+".") {
				return family.Name, true
			}
		}
	}

	return "", false
}

// ValidateDistribution checks if the CLI is being run on a supported OS family.
func ValidateDistribution(
	logger *slog.Logger,
) {
	info, err := hostInfoFn()
	if err != nil {
		LogFatal(logger, "failed to get host info", err)
	}

	if os.Getenv("IGNORE_LINUX") != "" {
		return
	}

	family, supported := IsOSFamilySupported(info.Platform, info.PlatformVersion)
	if !supported {
		LogFatal(
			logger,
			"os family not supported",
			fmt.Errorf(
				"%s %s is not a supported distribution",
				info.Platform,
				info.PlatformVersion,
			),
			"distro",
			info.Platform,
			"version",
			info.PlatformVersion,
		)
	}

	logger.Debug(
		"os family detected",
		slog.String("family", family),
		slog.String("distro", info.Platform),
		slog.String("version", info.PlatformVersion),
	)
}
