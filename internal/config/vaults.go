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

package config

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"

	"github.com/osapi-io/osapi/internal/secret/vault"
	// The nats type of vault, which the check of a vault knows.
	_ "github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// vaultsKey is the config key of the vaults.
const vaultsKey = "controller.secrets.vaults"

// vaultProblems are the problems of each vault of specs: two vaults with one
// name or one key file, a type that is not registered, a config block that
// its type does not take or whose values break their rules, and a spec that
// the check of its type refuses.
func vaultProblems(
	v *validator.Validate,
	specs []vault.Spec,
) []Problem {
	var problems []Problem
	add := func(i int, suffix string, msg string) {
		problems = append(problems, Problem{
			Rule:    RuleVault,
			Key:     fmt.Sprintf("%s[%d]%s", vaultsKey, i, suffix),
			Message: msg,
		})
	}
	names := map[string]int{}
	files := map[string]int{}
	for i, spec := range specs {
		if j, ok := names[spec.Name]; ok {
			add(
				i,
				".name",
				fmt.Sprintf("is %q, as vault %d is: each vault needs its own name", spec.Name, j),
			)
		}
		names[spec.Name] = i
		if spec.KeyFile != "" {
			if j, ok := files[spec.KeyFile]; ok {
				add(i, ".key_file", fmt.Sprintf(
					"is %q, as vault %d is: a rotation of one key would drop the key of the other vault",
					spec.KeyFile,
					j,
				))
			}
			files[spec.KeyFile] = i
		}
		t, err := vault.TypeOf(spec)
		if err != nil {
			add(i, ".type", err.Error())

			continue
		}
		cfg, err := t.Decode(spec)
		if err != nil {
			add(i, ".config", err.Error())

			continue
		}
		if err := v.Struct(cfg); err != nil {
			problems = append(problems, valueProblems(
				err, RuleVault, t.Config, fmt.Sprintf("%s[%d].config", vaultsKey, i),
			)...)
		}
		if err := t.Check(spec); err != nil {
			add(i, "", err.Error())
		}
	}

	return problems
}

// VaultKeys lists the settings of a vault, under controller.secrets.vaults[],
// then the settings of the config block of each registered type, under
// controller.secrets.vaults[].config, with the type in the doc. A table of
// the docs shows them, as it shows the keys of Keys. An environment variable
// does not set one, because a vault is an item of a list.
func VaultKeys() []Key {
	prefix := vaultsKey + "[]"
	keys := walk(reflect.TypeFor[vault.Spec](), prefix, reflect.Value{})
	for _, t := range vault.Types() {
		for _, k := range walk(t.Config, prefix+".config", reflect.Value{}) {
			k.Doc = fmt.Sprintf("Type %s: %s", t.Name, k.Doc)
			keys = append(keys, k)
		}
	}
	for i := range keys {
		keys[i].Env = ""
	}

	return keys
}
