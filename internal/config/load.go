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
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/validation"
)

// Where a key's value came from, highest precedence first.
const (
	SourceEnv     = "env"
	SourceFile    = "file"
	SourceDefault = "default"
	SourceUnset   = "unset"
)

// Load reads the config file at path into v and c, with every key's default
// and environment variable applied, then checks it. It is what osapi start
// runs: a missing file or a value that breaks a rule is an error.
func Load(
	v *viper.Viper,
	path string,
	c *Config,
) error {
	found, err := Read(v, path, c)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("read %s: %w", path, os.ErrNotExist)
	}
	if msg, ok := validation.Struct(*c); !ok {
		return errors.New(msg)
	}
	for _, b := range behavior.All() {
		if b.SettingsType() == nil {
			continue
		}
		if msg, ok := validation.Struct(behavior.Settings[any](b)); !ok {
			return fmt.Errorf("%s: %s", b.SettingsKey(), msg)
		}
	}

	return Validate(c)
}

// Read reads the config file at path into v and c, with every key's default
// and environment variable applied, without checking it. A missing file is
// not an error, so defaults and the environment can be shown without one;
// found says whether it was there. Every key is bound to its environment
// variable, so one set in the environment applies even when the file does not
// mention the key.
func Read(
	v *viper.Viper,
	path string,
	c *Config,
) (found bool, err error) {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetConfigType("yaml")
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()

	for _, k := range Keys() {
		if k.Default != nil {
			v.SetDefault(k.Name, k.Default)
		}
		_ = v.BindEnv(k.Name)
	}

	found = true
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("read %s: %w", path, err)
		}
		found = false
	}
	if err := v.Unmarshal(c); err != nil {
		return found, fmt.Errorf("parse %s: %w", path, err)
	}

	// Each behavior's settings are decoded into its own struct and handed to
	// it, where its methods read them. UnmarshalKey would read only the
	// file's part of the section, missing defaults and the environment, so
	// the section is rebuilt from each key's resolved value first.
	sections := map[string]*viper.Viper{}
	for _, k := range Keys() {
		if k.Behavior == "" {
			continue
		}
		b := behavior.Lookup(k.Behavior)
		section, ok := sections[k.Behavior]
		if !ok {
			section = viper.New()
			sections[k.Behavior] = section
		}
		if val := v.Get(k.Name); val != nil {
			section.Set(strings.TrimPrefix(k.Name, b.SettingsKey()+"."), val)
		}
	}
	for name, section := range sections {
		b := behavior.Lookup(name)
		settings := reflect.New(b.SettingsType())
		if err := section.Unmarshal(settings.Interface()); err != nil {
			return found, fmt.Errorf("parse %s: %s: %w", path, b.SettingsKey(), err)
		}
		b.SetSettings(settings.Elem().Interface())
	}

	return found, nil
}

// Source is where key's value in v came from: the environment, the config
// file, a default, or nowhere.
func Source(
	v *viper.Viper,
	key string,
) string {
	if _, ok := os.LookupEnv(EnvName(key)); ok {
		return SourceEnv
	}
	if v.InConfig(key) {
		return SourceFile
	}
	if _, ok := defaultOf(key); ok {
		return SourceDefault
	}

	return SourceUnset
}

// Value is k's value in v, as k's type: an environment variable arrives as
// text, and is converted so it reads the same as a value from the file.
func Value(
	v *viper.Viper,
	k Key,
) any {
	raw := v.Get(k.Name)
	if raw == nil {
		return nil
	}

	switch k.Type {
	case "int":
		return cast.ToInt64(raw)
	case "float":
		return cast.ToFloat64(raw)
	case "bool":
		return cast.ToBool(raw)
	default:
		return raw
	}
}
