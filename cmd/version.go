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

package cmd

import (
	goversion "github.com/caarlos0/go-version"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	versioncmd "github.com/osapi-io/osapi/internal/cli/commands/version"
)

func buildVersion(
	version, commit, date, builtBy, treeState string,
) goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails(
			"osapi",
			"Typed, repeatable operations for Linux and macOS hosts.\n",
			"https://github.com/osapi-io/osapi",
		),
		func(i *goversion.Info) {
			if commit != "" {
				i.GitCommit = commit
			}
			if treeState != "" {
				i.GitTreeState = treeState
			}
			if date != "" {
				i.BuildDate = date
			}
			if version != "" {
				i.GitVersion = version
			}
			if builtBy != "" {
				i.BuiltBy = builtBy
			}
		},
	)
}

// appVersion is the build version of this binary, such as
// v0.2.1-0.20261119003517-f6bc5af55ceb, which each component reports.
func appVersion() string {
	return buildVersion(version, commit, date, builtBy, treeState).GitVersion
}

// versionAttr is the version of this binary, for the resource of each span.
func versionAttr() attribute.KeyValue {
	return semconv.ServiceVersion(appVersion())
}

// The build of this binary, which the release sets with -ldflags.
var (
	version   = ""
	commit    = ""
	treeState = ""
	date      = ""
	builtBy   = ""
)

// buildInfo is the build of this binary, for osapi version.
func buildInfo() versioncmd.Info {
	v := buildVersion(version, commit, date, builtBy, treeState)

	return versioncmd.Info{
		Version:   v.GitVersion,
		Commit:    v.GitCommit,
		TreeState: v.GitTreeState,
		BuildDate: v.BuildDate,
		BuiltBy:   v.BuiltBy,
		GoVersion: v.GoVersion,
		Platform:  v.Platform,
		Compiler:  v.Compiler,
		ModuleSum: v.ModuleSum,
	}
}
