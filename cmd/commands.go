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
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/cli/commands/agentservice"
	"github.com/osapi-io/osapi/internal/cli/commands/pkicert"
	"github.com/osapi-io/osapi/internal/cli/commands/pkikey"
	"github.com/osapi-io/osapi/internal/cli/commands/pkilist"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/osapi-io/osapi/internal/controller/certrotation"
	"github.com/osapi-io/osapi/internal/controller/jointoken"

	sdkclient "github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/guide/sdktypes"

	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/cli/commands/chain"

	"github.com/avfs/avfs/vfs/osfs"

	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/collectors"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/agentcmd"

	// The generated command groups register themselves in their init.
	_ "github.com/osapi-io/osapi/internal/cli/commands/all/gen"
	"github.com/osapi-io/osapi/internal/cli/commands/behaviordsl"
	"github.com/osapi-io/osapi/internal/cli/commands/clidocs"
	"github.com/osapi-io/osapi/internal/cli/commands/commanddsl"
	configcmd "github.com/osapi-io/osapi/internal/cli/commands/config"
	"github.com/osapi-io/osapi/internal/cli/commands/controllercmd"
	"github.com/osapi-io/osapi/internal/cli/commands/endpointdsl"
	"github.com/osapi-io/osapi/internal/cli/commands/join"
	natsauthcmd "github.com/osapi-io/osapi/internal/cli/commands/natsauth"
	"github.com/osapi-io/osapi/internal/cli/commands/natsperms"
	"github.com/osapi-io/osapi/internal/cli/commands/natsserver"
	"github.com/osapi-io/osapi/internal/cli/commands/natssetup"
	"github.com/osapi-io/osapi/internal/cli/commands/role"
	"github.com/osapi-io/osapi/internal/cli/commands/sdk"
	"github.com/osapi-io/osapi/internal/cli/commands/skill"
	"github.com/osapi-io/osapi/internal/cli/commands/startcmd"
	"github.com/osapi-io/osapi/internal/cli/commands/token"
	versioncmd "github.com/osapi-io/osapi/internal/cli/commands/version"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
	"github.com/osapi-io/osapi/internal/skills"
)

// The commands of the local groups, which run in the CLI with no API, go
// under the root. Each group whose Deps has a field gets it here, made when
// a command runs, after the config is read.
func init() {
	command.Provide(token.Group, func() token.Deps {
		return token.Deps{Tokens: authtoken.New(logger), Config: tokenConfig{}, Audit: localAudit}
	})
	command.Provide(role.Group, func() role.Deps {
		return role.Deps{Config: tokenConfig{}}
	})
	command.Provide(configcmd.Group, func() configcmd.Deps {
		return configcmd.Deps{File: func() string { return viper.GetString("osapiFile") }}
	})
	command.Provide(join.Group, joinDeps)
	command.Provide(sdk.Group, func() sdk.Deps {
		return sdk.Deps{Types: types.All, Services: sdkclient.Services}
	})
	command.Provide(skill.Group, func() skill.Deps {
		return skill.Deps{
			Installed: skills.Installed(),
			Dev:       skills.Dev(),
			FS:        appFs,
			Version:   displayVersion,
		}
	})
	command.Provide(natsserver.Group, natsDeps)
	command.Provide(agentcmd.Group, agentDeps)
	command.Provide(controllercmd.Group, controllerDeps)
	command.Provide(natsauthcmd.Group, func() natsauthcmd.Deps {
		return natsauthcmd.Deps{FS: appFs, Config: appConfig, Audit: localAudit}
	})
	command.Provide(natsperms.Group, func() natsperms.Deps {
		return natsperms.Deps{Config: appConfig}
	})
	command.Provide(natssetup.Group, func() natssetup.Deps {
		return natssetup.Deps{
			Config:  appConfig,
			Logger:  logger,
			Prepare: bus.Prepare,
			Audit:   localAudit,
		}
	})
	command.Provide(pkikey.Group, pkiKeyDeps)
	command.Provide(pkilist.Group, func() pkilist.Deps {
		return pkilist.Deps{
			Keys: func() []pkikey.Key { return pkikey.Keys(pkiKeyDeps()) },
			Certs: func() []pki.Cert {
				return pki.ReadCerts(osfs.NewWithNoIdm(), config.CertSources(appConfig), time.Now())
			},
		}
	})
	command.Provide(pkicert.Group, func() pkicert.Deps {
		keys := pkiKeyDeps()

		return pkicert.Deps{
			FS:     osfs.NewWithNoIdm(),
			Certs:  func() []pki.CertSource { return config.CertSources(appConfig) },
			Now:    time.Now,
			Audit:  localAudit,
			Config: appConfig,
			Dial: func(ctx context.Context) (certrotation.KV, func(), error) {
				return keys.Dial(ctx)
			},
		}
	})
	command.Provide(agentservice.Group, func() agentservice.Deps {
		return agentservice.Deps{
			FS:         osfs.NewWithNoIdm(),
			GOOS:       runtime.GOOS,
			EUID:       os.Geteuid,
			Binary:     os.Executable,
			Run:        runCommand,
			ConfigFile: osapiFile(),
			KeyDir:     appConfig.Agent.PKI.KeyDir,
			Wait:       waitForFile,
		}
	})
	command.Provide(commanddsl.Group, commandDSLDeps)
	command.Provide(endpointdsl.Group, endpointDSLDeps)
	command.Provide(behaviordsl.Group, behaviorDSLDeps)
	command.Provide(chain.Group, chainDeps)
	command.Provide(clidocs.Group, func() clidocs.Deps {
		return clidocs.Deps{
			Root:   rootCmd,
			DSLs:   dslPages,
			Pages:  referencePages,
			Tables: configTables,
		}
	})
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
	}, natsCmd, agentCmd, controllerCmd, dslCmd, pkiCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// tokenConfig is the config that the token and role groups read.
type tokenConfig struct{}

// SigningKeys are controller.api.security.signing_keys of the config that
// osapi read at startup.
func (tokenConfig) SigningKeys(
	ctx context.Context,
) (authtoken.Keys, error) {
	if err := resolveReferences(ctx, "controller.api.security.signing_keys"); err != nil {
		return nil, err
	}
	if err := config.EnsureSigningKey(&appConfig); err != nil {
		return nil, err
	}

	return appConfig.Controller.API.Security.Keys(), nil
}

// CustomRoles are the roles of the config file, read here: role list
// needs no config file, so osapi does not read one at startup for it.
func (tokenConfig) CustomRoles() (authtoken.CustomRoles, error) {
	var c config.Config
	if _, err := config.Read(viper.New(), viper.GetString("osapiFile"), &c); err != nil {
		return nil, err
	}
	out := make(authtoken.CustomRoles, len(c.Controller.API.Security.Roles))
	for name, r := range c.Controller.API.Security.Roles {
		out[name] = r.Permissions
	}

	return out, nil
}

// dslPages are the DSLs of the binary as pages of the docs site, each from the
// spec that its spec get command prints. osapi dsl docs generate writes
// them.
func dslPages() []clidocs.DSL {
	return []clidocs.DSL{
		{
			Slug:     "behavior",
			Title:    "Behavior DSL",
			Position: 2,
			Intro: "Each part of a behavior definition: what it makes osapi do, and when to use it. " +
				"`osapi dsl behavior spec get` prints the same list from this build. " +
				"[Building a behavior](../behaviors.md) tells how to write one.",
			Spec: sdktypes.Append(collectors.Append(behavior.CurrentSpec(job.FactKeys()...))),
		},
		{
			Slug: "chain", Title: "Chains", Position: 3,
			Intro: "How a behavior builds on others and calls them on the same host, what occurs " +
				"when a step fails, and the building blocks of this build. " +
				"`osapi dsl chain spec get` prints the same list.",
			Spec: behavior.CurrentChainSpec(),
		},
		{
			Slug: "endpoint", Title: "Endpoint DSL", Position: 4,
			Intro: "Each part of an endpoint group: a route of the controller, its spec and its " +
				"commands. `osapi dsl endpoint spec get` prints the same list from this build. " +
				"[Building an endpoint](../endpoints.md) tells how to write one.",
			Spec: sdktypes.Append(endpoint.Spec),
		},
		{
			Slug: "command", Title: "Command DSL", Position: 5,
			Intro: "Each part of a local group: a command that runs in the CLI and calls no API. " +
				"`osapi dsl command spec get` prints the same list from this build. " +
				"[Building a command](../commands.md) tells how to write one.",
			Spec: sdktypes.Append(command.Spec),
		},
	}
}

// commandDSLDeps is what the commanddsl commands use, and the reference pages too.
func commandDSLDeps() commanddsl.Deps {
	return commanddsl.Deps{
		Root:       rootCmd,
		Dir:        ".",
		IsBehavior: func(name string) bool { return behavior.Lookup(name) != nil },
	}
}

// endpointDSLDeps is what the endpointdsl commands use, and the reference pages too.
func endpointDSLDeps() endpointdsl.Deps {
	return endpointdsl.Deps{
		Groups: endpoint.All,
		OpenAPI: func(g *endpoint.Group) (any, error) {
			return endpoint.OpenAPI(g)
		},
		Dir: ".",
	}
}

// behaviorDSLDeps is what the behaviordsl commands use, and the reference pages too.
func behaviorDSLDeps() behaviordsl.Deps {
	return behaviordsl.Deps{
		Behaviors: behavior.All,
		Validate:  behavior.Validate,
		Keys:      config.Keys,
		Spec:      func() guide.Spec { return sdktypes.Append(collectors.Append(behavior.CurrentSpec(job.FactKeys()...))) },
		OpenAPI: func(b *behavior.Behavior) (any, error) {
			g := behavior.GroupOf(b)
			if g == nil {
				return nil, fmt.Errorf("%s has no endpoint group: run go generate", b.Name)
			}

			return endpoint.OpenAPI(g)
		},
	}
}

// chainDeps is what the chain commands use, and the reference pages too.
func chainDeps() chain.Deps {
	return chain.Deps{
		Behaviors:   behavior.All,
		Runs:        behaviortest.Chains,
		Problems:    behavior.ChainProblems,
		RunProblems: behaviortest.ChainProblems,
		Spec:        behavior.CurrentChainSpec,
	}
}

// joinDeps is what osapi controller join token uses: the config, a
// connection to NATS as the controller makes it, and the time.
func joinDeps() join.Deps {
	return join.Deps{
		Config: appConfig,
		Now:    time.Now,
		Dial: func(context.Context) (jointoken.KV, func(), error) {
			n := appConfig.Controller.NATS
			opts, err := bus.ConnectionOptions(n, "osapi-join-token")
			if err != nil {
				return nil, nil, err
			}
			nc, err := bus.Dial(logger, opts)
			if err != nil {
				return nil, nil, err
			}

			return nc, func() { bus.Close(nc) }, nil
		},
		Audit: localAudit,
	}
}

// controllerInFile reports whether the config file sets the controller. An
// agent host's file sets only agent: its join gives it the agent's NATS
// connection, and no controller.nats to write an audit entry with
// (osapi-io/osapi#1552). A test replaces it.
var controllerInFile = func() bool { return viper.InConfig("controller") }

// localAudit writes e to the audit log of the controller, with a connection
// to NATS as the controller makes it. The controller keeps no audit log when
// nats.audit.stream is empty. An agent host keeps no local audit log: the
// controller audits what reaches it, such as agent.key.rotate.
func localAudit(
	ctx context.Context,
	e audit.Entry,
) error {
	if !controllerInFile() {
		logger.Info("an agent host keeps no local audit log: the controller audits what reaches it",
			slog.String("action", e.Action),
		)

		return nil
	}
	err := writeLocalAudit(ctx, e)
	if err != nil {
		// The change was made. Say that the audit log does not have it,
		// such as for nats auth create before NATS starts.
		logger.Warn("the change is not in the audit log: NATS could not be reached",
			slog.String("action", e.Action),
			slog.String("error", err.Error()),
		)
	}

	return err
}

// writeLocalAudit is localAudit without the warning.
func writeLocalAudit(
	ctx context.Context,
	e audit.Entry,
) error {
	a := appConfig.NATS.Audit
	if a.Stream == "" {
		return nil
	}
	n := appConfig.Controller.NATS
	opts, err := bus.ConnectionOptions(n, "osapi-audit")
	if err != nil {
		return err
	}
	nc, err := bus.Dial(logger, opts)
	if err != nil {
		return err
	}
	defer bus.Close(nc)
	store := audit.NewStreamStore(logger, nc,
		job.ApplyNamespaceToInfraName(n.Namespace, a.Stream),
		job.ApplyNamespaceToSubjects(n.Namespace, a.Subject))

	return store.Write(ctx, e)
}

// pkiKeyDeps is what osapi pki key uses: the file system, the config, the
// identity of this machine, and a connection to NATS as the controller makes
// it, for the rotation of a vault key.
func pkiKeyDeps() pkikey.Deps {
	fs := osfs.NewWithNoIdm()

	return pkikey.Deps{
		Audit:     localAudit,
		FS:        fs,
		Config:    appConfig,
		LookupEnv: os.LookupEnv,
		Identity: func() (*identity.Identity, error) {
			return identity.GetIdentity(fs, appConfig.Agent.Hostname)
		},
		RotateAgent: func(ctx context.Context) (string, error) {
			return agent.RotateKey(ctx, logger, fs, appConfig)
		},
		Dial: func(context.Context) (natsvault.KV, func(), error) {
			n := appConfig.Controller.NATS
			opts, err := bus.ConnectionOptions(n, "osapi-vault-key")
			if err != nil {
				return nil, nil, err
			}
			nc, err := bus.Dial(logger, opts)
			if err != nil {
				return nil, nil, err
			}

			return nc, func() { bus.Close(nc) }, nil
		},
	}
}

// runCommand runs name with args, and returns its output in the error when
// it fails.
func runCommand(
	ctx context.Context,
	name string,
	args ...string,
) error {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

// osapiFile is the config file of this run: -f, or the default path.
func osapiFile() string {
	if f := viper.GetString("osapiFile"); f != "" {
		return f
	}

	return config.DefaultFile
}

// waitForFile is whether file exists within a minute, or before ctx ends.
func waitForFile(
	ctx context.Context,
	file string,
) bool {
	deadline := time.After(time.Minute)
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	for {
		if _, err := os.Stat(file); err == nil {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-deadline:
			return false
		case <-tick.C:
		}
	}
}
