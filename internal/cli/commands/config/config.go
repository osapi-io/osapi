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

// Package token is the token group of local commands: osapi controller token generate,
// roles list and validate, which sign and read the bearer tokens of the API
// with the signing key of the config. They run in the CLI, and call no API.
// Package config is the config group: osapi config list, get and validate,
// which read osapi's settings the way osapi start does.
package config

import (
	"github.com/osapi-io/osapi/internal/cli/command"
	osapiconfig "github.com/osapi-io/osapi/internal/config"
)

// Group is the config group's definition.
var Group = command.Register(command.Define(
	"config",
	"List, get and validate osapi's settings",
	command.Long(
		"These commands load the config themselves. Thus they report a broken file,\n"+
			"and the command does not stop.",
	),
	command.Uses[Deps](),
	command.Method[ListInput, ListOutput]("list").
		Short("List every setting, its value in effect and where it came from").
		Long("Each value shows where it came from: `env`, `file`, `default` or `unset`.\n"+
			"The file and the environment are read the way `osapi start` reads them.\n"+
			"Without a config file, the values are the defaults and the environment.\n"+
			"Passwords, signing keys and tokens print as `********`. `--json` carries the\n"+
			"type, the default, the rules and the description of each key as well.").
		Example("").
		Prints(sampleList).
		Example("-f ./osapi.yaml --json").
		Handle(List),
	command.Method[GetInput, GetOutput]("get").
		Short("Show one setting in full").
		Long("It shows the value in effect and where it came from, the environment\n"+
			"variable, the type, the default, the rules and the description. An\n"+
			"environment variable wins over the file.").
		Example("--key controller.api.port").
		Prints(sampleGet).
		Example("--key agent.labels -f ./osapi.yaml --json").
		Fails("no key", GetInput{}).
		Handle(Get),
	command.Method[ValidateInput, ValidateOutput]("validate").
		Short("Check a config file before osapi starts with it").
		Long("The command starts nothing. It exits 1 when the config is not valid, and\n"+
			"lists each problem with its rule, the key and the message, then the fix of\n"+
			"each rule. Each problem in `--json` has `rule`, `key` (when the problem is\n"+
			"about one key), `message` and `fix`. `osapi config list` shows each key,\n"+
			"with its type, rule and default.\n\nThe rules:\n\n"+osapiconfig.Checks.Help()).
		Example("").
		Prints(sampleValidate).
		Example("-f ./osapi.yaml --json").
		Handle(Validate),
))
