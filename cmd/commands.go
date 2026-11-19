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
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all/gen"
	"github.com/osapi-io/osapi/internal/cli/commands/token"
	"github.com/osapi-io/osapi/internal/config"
)

// The commands of the local groups, which run in the CLI with no API, go
// under the root. Each group whose Deps has a field gets it here, made when
// a command runs, after the config is read.
func init() {
	command.Provide(token.Group, func() token.Deps {
		return token.Deps{Tokens: authtoken.New(logger), Config: tokenConfig{}}
	})

	if err := command.AttachLocal(rootCmd, command.Deps{
		JSON:      func() bool { return jsonOutput },
		Error:     func(err error) { cli.HandleError(err, logger) },
		PrintJSON: cli.PrintJSON,
		Print:     cli.Printer{},
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// tokenConfig is the config the token group reads.
type tokenConfig struct{}

// SigningKey is controller.api.security.signing_key of the config that
// osapi read at startup.
func (tokenConfig) SigningKey() string {
	return appConfig.Controller.API.Security.SigningKey
}

// CustomRoles are the roles of the config file, read here: roles list
// needs no config file, so osapi does not read one at startup for it.
func (tokenConfig) CustomRoles() (map[string][]string, error) {
	var c config.Config
	if _, err := config.Read(viper.New(), viper.GetString("osapiFile"), &c); err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(c.Controller.API.Security.Roles))
	for name, r := range c.Controller.API.Security.Roles {
		out[name] = r.Permissions
	}

	return out, nil
}
