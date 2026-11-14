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
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// roleDescription is one role in `token roles --json`.
type roleDescription struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Permissions []string `json:"permissions"`
}

var tokenRolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "List the roles and the permissions each grants",
	Long: `Lists the built-in roles, admin, write and read, and any custom role in
the config, with every permission each grants. A behavior's permissions,
<name>:read and <name>:write, are included: admin and write grant both, read
grants read. Works without a config file, showing the built-in roles only.`,
	Example: `  osapi token roles
  osapi token roles --json`,
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		var c config.Config
		if _, err := config.Read(viper.New(), viper.GetString("osapiFile"), &c); err != nil {
			return err
		}

		roles := builtinRoles()
		custom := make([]string, 0, len(c.Controller.API.Security.Roles))
		for name := range c.Controller.API.Security.Roles {
			custom = append(custom, name)
		}
		sort.Strings(custom)
		for _, name := range custom {
			roles = append(roles, roleDescription{
				Name:        name,
				Source:      "config",
				Permissions: c.Controller.API.Security.Roles[name].Permissions,
			})
		}

		if jsonOutput {
			out, _ := json.Marshal(map[string]any{"roles": roles})
			cli.PrintJSON(out)

			return nil
		}

		printRoles(roles)

		return nil
	},
}

// builtinRoles are admin, write and read, with what each grants.
func builtinRoles() []roleDescription {
	roles := make([]roleDescription, 0, 3)
	for _, name := range []string{client.RoleAdmin, client.RoleWrite, client.RoleRead} {
		perms, _ := authtoken.RolePermissions(name)
		roles = append(roles, roleDescription{Name: name, Source: "built-in", Permissions: perms})
	}

	return roles
}

// printRoles prints one row per permission, with a column per role saying
// whether it grants it.
func printRoles(
	roles []roleDescription,
) {
	headers := make([]string, 0, 1+len(roles))
	headers = append(headers, "PERMISSION")
	for _, r := range roles {
		headers = append(headers, r.Name)
	}

	perms := authtoken.AllPermissions()
	rows := make([][]string, 0, len(perms))
	for _, p := range perms {
		row := make([]string, 0, 1+len(roles))
		row = append(row, p)
		for _, r := range roles {
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

func init() {
	tokenCmd.AddCommand(tokenRolesCmd)
}
