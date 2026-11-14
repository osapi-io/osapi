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
)

// configCmd groups the commands that describe osapi's own settings. They load
// the config themselves, so a broken file is reported rather than fatal.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "List, get and validate osapi's settings",
	Long: `osapi reads its settings from a YAML file (--osapi-file, default
/etc/osapi/osapi.yaml), then from OSAPI_ environment variables, which win over
the file, then from built-in defaults. These commands list every setting with the
value in effect and where it came from, show one in full, and check a file
before osapi starts with it.`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
