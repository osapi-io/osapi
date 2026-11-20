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

	"github.com/avfs/avfs/vfs/osfs"

	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/admission"
	"github.com/osapi-io/osapi/internal/cli/commands/agentcmd"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all/gen"
	"github.com/osapi-io/osapi/internal/cli/commands/behaviordsl"
	"github.com/osapi-io/osapi/internal/cli/commands/clidocs"
	"github.com/osapi-io/osapi/internal/cli/commands/commanddsl"
	configcmd "github.com/osapi-io/osapi/internal/cli/commands/config"
	"github.com/osapi-io/osapi/internal/cli/commands/controllercmd"
	"github.com/osapi-io/osapi/internal/cli/commands/controllerkey"
	"github.com/osapi-io/osapi/internal/cli/commands/endpointdsl"
	"github.com/osapi-io/osapi/internal/cli/commands/initcmd"
	"github.com/osapi-io/osapi/internal/cli/commands/natsserver"
	"github.com/osapi-io/osapi/internal/cli/commands/natssetup"
	"github.com/osapi-io/osapi/internal/cli/commands/skill"
	"github.com/osapi-io/osapi/internal/cli/commands/startcmd"
	"github.com/osapi-io/osapi/internal/cli/commands/token"
	versioncmd "github.com/osapi-io/osapi/internal/cli/commands/version"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/skills"
)

// The commands of the local groups, which run in the CLI with no API, go
// under the root. Each group whose Deps has a field gets it here, made when
// a command runs, after the config is read.
func init() {
	command.Provide(token.Group, func() token.Deps {
		return token.Deps{Tokens: authtoken.New(logger), Config: tokenConfig{}}
	})
	command.Provide(configcmd.Group, func() configcmd.Deps {
		return configcmd.Deps{File: func() string { return viper.GetString("osapiFile") }}
	})
	command.Provide(skill.Group, func() skill.Deps {
		return skill.Deps{Installed: skills.Installed(), Dev: skills.Dev()}
	})
	command.Provide(initcmd.Group, func() initcmd.Deps {
		return initcmd.Deps{FS: appFs, Source: skills.Installed(), Version: displayVersion}
	})
	command.Provide(natsserver.Group, natsDeps)
	command.Provide(agentcmd.Group, agentDeps)
	command.Provide(controllercmd.Group, controllerDeps)
	command.Provide(controllerkey.Group, func() controllerkey.Deps {
		return controllerkey.Deps{FS: osfs.NewWithNoIdm(), KeyDir: appConfig.Controller.PKI.KeyDir}
	})
	command.Provide(natssetup.Group, func() natssetup.Deps {
		return natssetup.Deps{Config: appConfig, Logger: logger, Prepare: bus.Prepare}
	})
	command.Provide(admission.Group, func() admission.Deps {
		fs := osfs.NewWithNoIdm()

		return admission.Deps{
			FS:     fs,
			KeyDir: appConfig.Agent.PKI.KeyDir,
			Identity: func() (*identity.Identity, error) {
				return identity.GetIdentity(fs, appConfig.Agent.Hostname)
			},
		}
	})
	command.Provide(commanddsl.Group, func() commanddsl.Deps {
		return commanddsl.Deps{
			Root:       rootCmd,
			Dir:        ".",
			IsBehavior: func(name string) bool { return behavior.Lookup(name) != nil },
		}
	})
	command.Provide(endpointdsl.Group, func() endpointdsl.Deps {
		return endpointdsl.Deps{
			Groups: endpoint.All,
			OpenAPI: func(g *endpoint.Group) (any, error) {
				return endpoint.OpenAPI(g, endpoint.ErrorRef)
			},
			Dir: ".",
		}
	})
	command.Provide(behaviordsl.Group, func() behaviordsl.Deps {
		return behaviordsl.Deps{
			Behaviors: behavior.All,
			Validate:  behavior.Validate,
			Keys:      config.Keys,
			Spec:      func() guide.Spec { return behavior.CurrentSpec(job.FactKeys()...) },
		}
	})
	command.Provide(clidocs.Group, func() clidocs.Deps { return clidocs.Deps{Root: rootCmd} })
	command.Provide(startcmd.Group, func() startcmd.Deps {
		return startcmd.Deps{NATS: natsDeps(), Controller: controllerDeps(), Agent: agentDeps()}
	})
	command.Provide(versioncmd.Group, func() versioncmd.Deps {
		return versioncmd.Deps{Build: buildInfo}
	})

	if err := command.AttachLocal(rootCmd, command.Deps{
		JSON:      func() bool { return jsonOutput },
		Error:     func(err error) { cli.HandleError(err, logger) },
		PrintJSON: cli.PrintJSON,
		Print:     cli.Printer{},
		// These are written by hand, and their inits run after this one: a
		// group goes Under them.
	}, natsCmd, agentCmd, controllerCmd, cliCmd); err != nil {
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
