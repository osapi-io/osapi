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

package token

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"
)

func init() {
	command.Render(Group, "generate", renderGenerate)
	command.Render(Group, "list", renderRoles)
	command.Render(Group, "validate", renderClaims)
}

// renderGenerate prints the token, and what it carries.
func renderGenerate(
	_ *cobra.Command,
	_ GenerateInput,
	out GenerateOutput,
) {
	fmt.Println()
	cli.PrintKV("Token", out.Token)
	cli.PrintKV("Subject", out.Subject, "Roles", strings.Join(out.Roles, ", "))
	if len(out.Permissions) > 0 {
		cli.PrintKV("Permissions", strings.Join(out.Permissions, ", "))
	}
}

// renderRoles prints one row for each permission, with a column for each
// role that says whether the role grants it.
func renderRoles(
	_ *cobra.Command,
	_ RolesInput,
	out RolesOutput,
) {
	headers := make([]string, 0, 1+len(out.Roles))
	headers = append(headers, "PERMISSION")
	for _, r := range out.Roles {
		headers = append(headers, r.Name)
	}

	perms := authtoken.AllPermissions()
	rows := make([][]string, 0, len(perms))
	for _, p := range perms {
		row := make([]string, 0, 1+len(out.Roles))
		row = append(row, p)
		for _, r := range out.Roles {
			mark := ""
			if slices.Contains(r.Permissions, p) {
				mark = "yes"
			}
			row = append(row, mark)
		}
		rows = append(rows, row)
	}

	cli.PrintCompactTable([]cli.Section{{
		Title:     fmt.Sprintf("Permissions (%d)", len(perms)),
		Headers:   headers,
		Rows:      rows,
		Reference: true,
	}})
}

// renderClaims prints what the token carries, and when it was issued and
// expires.
func renderClaims(
	_ *cobra.Command,
	_ ValidateInput,
	out Claims,
) {
	cli.PrintHeading("Token")
	cli.PrintKV("Subject", out.Subject, "Roles", strings.Join(out.Roles, ", "))
	if len(out.Permissions) > 0 {
		cli.PrintKV("Permissions", strings.Join(out.Permissions, ", "))
	}
	cli.PrintKV("Effective permissions", strings.Join(out.EffectivePermissions, ", "))
	cli.PrintKV("Audience", strings.Join(out.Audience, ", "))
	cli.PrintKV(
		"Issued", cli.LocalTime(out.IssuedAt),
		"Expires", cli.LocalTime(out.ExpiresAt),
	)
}
