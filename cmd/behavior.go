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

package cmd

import (
	"github.com/spf13/cobra"

	// Registers every behavior, so the commands below can see them.
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
)

// behaviorCmd groups the commands that read the behavior registry. They need
// no config and no controller: they describe what this binary can do.
var behaviorCmd = &cobra.Command{
	Use:   "behavior",
	Short: "List, describe and validate the behaviors this binary has",
	Long: `A behavior is what osapi can do to a host, such as manage kernel
parameters. Each one is a typed definition, and its API routes, CLI commands,
spec and docs are built from it. These commands read those definitions.

A behavior can be built on others. One that declares behavior.Requires calls
the required behaviors' methods directly, on the same host inside the same
job; 'osapi behavior get' lists them under Uses, and the behaviors built on
it under Used by. A behavior that exists only to be built on is internal: it
has no route, command or permission, and only other behaviors call it.`,
}

func init() {
	rootCmd.AddCommand(behaviorCmd)
	skipConfigCmds["behavior"] = true
}
