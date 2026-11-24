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
	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/behavior/command"
)

// agentCmd is the agent's command. Its start and its key are groups of the
// command DSL.
var agentCmd = &cobra.Command{
	Use: "agent",
	// Its help is its purpose in command.RootGroups, which DescribeRoot gives it.
	Long: "The agent runs on each managed host, processes jobs, and reports status back\n" +
		"to the control plane.",
}

func init() {
	rootCmd.AddCommand(agentCmd)

	// Each behavior's settings, as --<behavior>-<setting> flags bound to
	// behaviors.<behavior>.<setting>. The agent runs the methods that read
	// them, so the flags are here.
	command.SettingFlags(agentCmd.PersistentFlags(), viper.GetViper())
}
