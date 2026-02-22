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
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/authtoken"
)

// TokenGenerator generates signed JWT tokens.
type TokenGenerator interface {
	Generate(
		signingKey string,
		roles []string,
		subject string,
		permissions []string,
	) (string, error)
}

// tokenGenerateCmd represents the tokenGenerate command.
var tokenGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new token",
	Long: `Generate a new API token with specific roles, expiration, and issuer details.
This command allows you to customize the token properties for various use cases.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		signingKey := appConfig.API.Server.Security.SigningKey
		roles, _ := cmd.Flags().GetStringSlice("roles")
		subject, _ := cmd.Flags().GetString("subject")
		permissions, _ := cmd.Flags().GetStringSlice("permissions")

		var tm TokenGenerator = authtoken.New(logger)
		tokin, err := tm.Generate(signingKey, roles, subject, permissions)
		if err != nil {
			logFatal("failed to generate token", err)
		}

		logger.Info(
			"generated token",
			slog.String("token", tokin),
			slog.String("roles", strings.Join(roles, ",")),
			slog.String("subject", subject),
		)
		if len(permissions) > 0 {
			logger.Info(
				"token permissions",
				slog.String("permissions", strings.Join(permissions, ",")),
			)
		}
	},
}

func init() {
	tokenCmd.AddCommand(tokenGenerateCmd)
	allowedRoles := authtoken.GenerateAllowedRoles(authtoken.RoleHierarchy)
	usage := fmt.Sprintf("Roles for the token (allowed: %s)", strings.Join(allowedRoles, ", "))

	tokenGenerateCmd.PersistentFlags().
		StringSliceP("roles", "r", []string{}, usage)
	tokenGenerateCmd.PersistentFlags().
		StringP("subject", "u", "", "Subject for the token (e.g., user ID or unique identifier)")
	tokenGenerateCmd.PersistentFlags().
		StringSliceP("permissions", "p", []string{},
			fmt.Sprintf("Direct permissions (overrides role expansion; allowed: %s)",
				strings.Join(authtoken.AllPermissions, ", ")))

	_ = tokenGenerateCmd.MarkPersistentFlagRequired("roles")
	_ = tokenGenerateCmd.MarkPersistentFlagRequired("subject")

	tokenGenerateCmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		roles, _ := cmd.Flags().GetStringSlice("roles")

		if err := validateRoles(roles); err != nil {
			logFatal("invalid roles", err, "allowed", allowedRoles)
		}

		permissions, _ := cmd.Flags().GetStringSlice("permissions")
		if err := validatePermissions(permissions); err != nil {
			logFatal("invalid permissions", err, "allowed", authtoken.AllPermissions)
		}
	}
}

func validateRoles(
	roles []string,
) error {
	allowedRoles := authtoken.GenerateAllowedRoles(authtoken.RoleHierarchy)
	allowedRolesMap := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowedRolesMap[role] = struct{}{}
	}

	for _, role := range roles {
		if _, ok := allowedRolesMap[role]; !ok {
			return fmt.Errorf("unsupported role: %s", role)
		}
	}
	return nil
}

func validatePermissions(
	permissions []string,
) error {
	if len(permissions) == 0 {
		return nil
	}

	allowedMap := make(map[string]struct{}, len(authtoken.AllPermissions))
	for _, p := range authtoken.AllPermissions {
		allowedMap[p] = struct{}{}
	}

	for _, p := range permissions {
		if _, ok := allowedMap[p]; !ok {
			return fmt.Errorf("unsupported permission: %s", p)
		}
	}
	return nil
}
