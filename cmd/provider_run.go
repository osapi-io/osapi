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
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var providerRunData string

var providerRunCmd = &cobra.Command{
	Use:    "run [provider] [operation]",
	Short:  "Run a provider operation (internal)",
	Hidden: true,
	Args:   cobra.ExactArgs(2),
	RunE: func(
		_ *cobra.Command,
		args []string,
	) error {
		providerName := args[0]
		operationName := args[1]

		reg := buildProviderRegistry()
		spec, ok := reg.Lookup(providerName, operationName)
		if !ok {
			return fmt.Errorf("unknown provider/operation: %s/%s", providerName, operationName)
		}

		var params any
		if spec.NewParams != nil {
			params = spec.NewParams()
		}
		if providerRunData != "" && params != nil {
			if err := json.Unmarshal([]byte(providerRunData), params); err != nil {
				return fmt.Errorf("parse input data: %w", err)
			}
		}

		result, err := spec.Run(context.Background(), params)
		if err != nil {
			return err
		}

		output, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}

		_, _ = fmt.Fprintln(os.Stdout, string(output))

		return nil
	},
}

var providerCmd = &cobra.Command{
	Use:    "provider",
	Short:  "Provider operations (internal)",
	Hidden: true,
}

func init() {
	providerRunCmd.Flags().StringVar(&providerRunData, "data", "", "JSON input data")
	providerCmd.AddCommand(providerRunCmd)
	rootCmd.AddCommand(providerCmd)
}
