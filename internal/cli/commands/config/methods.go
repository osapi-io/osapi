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
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/cli/command"
	osapiconfig "github.com/osapi-io/osapi/internal/config"
)

// redacted stands in for a secret's value.
const redacted = "********"

// List is each setting with its value in effect and where it came from, read
// the way osapi start reads them. A missing file is not an error: File says
// so, and the values are the defaults and the environment.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	file, settings, err := read(d.File())
	if err != nil {
		return ListOutput{}, err
	}

	return ListOutput{File: file, Items: settings}, nil
}

// Get is one setting in full.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (GetOutput, error) {
	file, settings, err := read(d.File())
	if err != nil {
		return GetOutput{}, err
	}
	for _, s := range settings {
		if s.Name == in.Key {
			return GetOutput{File: file, Setting: s}, nil
		}
	}

	return GetOutput{}, fmt.Errorf("no setting %q: osapi config list shows them all", in.Key)
}

// Validate checks the config file with each rule. An invalid config is its
// output and command.Fail, so the command exits 1 after it prints.
func Validate(
	_ context.Context,
	d Deps,
	_ ValidateInput,
) (ValidateOutput, error) {
	file := d.File()
	var c osapiconfig.Config
	problems := osapiconfig.Check(viper.New(), file, &c)
	out := ValidateOutput{File: file, Valid: len(problems) == 0, Problems: problems}
	if out.Problems == nil {
		out.Problems = []osapiconfig.Problem{}
	}
	if out.Valid {
		return out, nil
	}
	out.Fixes = map[string]any{}
	for _, p := range problems {
		out.Fixes[p.Rule] = p.Fix
	}

	return out, command.Fail()
}

// read reads the config file and the environment without checking them, and
// is each key with its value and where it came from.
func read(
	file string,
) (string, []Setting, error) {
	v := viper.New()
	var c osapiconfig.Config
	found, err := osapiconfig.Read(v, file, &c)
	if err != nil {
		return file, nil, err
	}
	if !found {
		file += " (not found: defaults and environment only)"
	}

	settings := make([]Setting, 0, len(osapiconfig.Keys()))
	for _, k := range osapiconfig.Keys() {
		val := osapiconfig.Value(v, k)
		if k.Secret && val != nil && val != "" {
			val = redacted
		}
		s := Setting{
			Name: k.Name, Value: val, Source: osapiconfig.Source(v, k.Name), Env: k.Env,
			Type: k.Type, Default: k.Default, Validate: k.Validate, Doc: k.Doc,
			Behavior: k.Behavior, Secret: k.Secret,
		}
		if k.Flag != "" {
			s.Flag = "osapi agent start --" + k.Flag
		}
		settings = append(settings, s)
	}

	return file, settings, nil
}
