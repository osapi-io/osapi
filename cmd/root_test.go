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
	"os"
	"strings"
	"testing"
)

// TestSkipConfig: a command that reads no config is matched by its path,
// so osapi client agent config, a command of the API, reads the config and
// osapi config does not.
func TestSkipConfig(t *testing.T) {
	was := os.Args
	defer func() { os.Args = was }()
	for args, want := range map[string]bool{
		"version":                      true,
		"config list":                  true,
		"controller role list":         true,
		"client agent config update":   false,
		"client agent list":            false,
		"controller join token create": false,
	} {
		os.Args = append([]string{"osapi"}, strings.Fields(args)...)
		if got := skipConfig(); got != want {
			t.Errorf("osapi %s: skipConfig is %v, want %v", args, got, want)
		}
	}
}
