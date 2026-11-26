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
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/internal/settings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// EnvPrefix is the prefix of every environment variable that sets a key.
const EnvPrefix = "OSAPI"

// DefaultVaultKeyFile is the key file of the default vault, which the
// controller makes when it is not there.
const DefaultVaultKeyFile = "/etc/osapi/secrets.key"

// Key is one setting osapi reads: its dotted name, the environment variable
// that sets it, its type, its default, its constraints and what it is for.
type Key struct {
	Name    string `json:"name"`
	Env     string `json:"env"`
	Type    string `json:"type"`
	Default any    `json:"default,omitempty"`
	// Choice says why the key has no default: it is a real choice of the
	// operator, such as a URL or a credential.
	Choice   string `json:"choice,omitempty"`
	Validate string `json:"validate,omitempty"`
	Doc      string `json:"doc,omitempty"`
	// Secret is true for a value that is never printed: a password, a
	// signing key or a token.
	Secret bool `json:"secret,omitempty"`
	// Behavior names the behavior that declared the key, for a behavior
	// setting.
	Behavior string `json:"behavior,omitempty"`
	// Group names the endpoint group that declared the key, for a group
	// setting.
	Group string `json:"group,omitempty"`
	// Flag is the agent start flag bound to a behavior setting.
	Flag string `json:"flag,omitempty"`
}

// EnvName is the environment variable that sets key: OSAPI_, then the key in
// capitals with its dots as underscores.
func EnvName(
	key string,
) string {
	return EnvPrefix + "_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}

// Keys lists every setting: osapi's own, in the order the config types
// declare them, with each registered endpoint group's after the last key of
// the controller, under controller.endpoints.<name>, then each registered
// behavior's, under behaviors.<name>. A struct is walked into; a list, a map
// or a scalar is one key.
func Keys() []Key {
	keys := withGroups(walk(reflect.TypeFor[Config](), "", ownedDefaults()))
	for _, b := range behavior.All() {
		t := b.SettingsType()
		if t == nil {
			continue
		}
		prefix := b.SettingsKey()
		for _, k := range walk(t, prefix, reflect.ValueOf(b.Defaults)) {
			k.Behavior = b.Name
			k.Flag = b.SettingFlag(strings.TrimPrefix(k.Name, prefix+"."))
			keys = append(keys, k)
		}
	}

	return keys
}

// withGroups is own, osapi's own keys, with each endpoint group's settings
// after the last key under controller, so a file that lists the keys in
// order has one controller section.
func withGroups(
	own []Key,
) []Key {
	var groups []Key
	for _, g := range endpoint.All() {
		t := g.SettingsType()
		if t == nil {
			continue
		}
		for _, k := range walk(t, g.SettingsKey(), reflect.ValueOf(g.Defaults)) {
			k.Group = g.Name
			groups = append(groups, k)
		}
	}
	at := len(own)
	for i, k := range own {
		if strings.HasPrefix(k.Name, "controller.") {
			at = i + 1
		}
	}

	return append(own[:at:at], append(groups, own[at:]...)...)
}

// ownedDefaults is a Config whose sections are the defaults that their
// owners declare (settings.Own, through command.Config): the source of the
// default of each key of osapi's own config.
func ownedDefaults() reflect.Value {
	root := reflect.New(reflect.TypeFor[Config]()).Elem()
	t := root.Type()
	for i := range t.NumField() {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("mapstructure"), ",")
		if o, ok := settings.Of(name); ok {
			root.Field(i).Set(reflect.ValueOf(o.Defaults))
		}
	}

	return root
}

// defaultOf is key's default, from the definition that owns it.
func defaultOf(
	key string,
) (any, bool) {
	for _, k := range Keys() {
		if k.Name == key {
			return k.Default, k.Default != nil
		}
	}

	return nil, false
}

// walk lists the keys of struct type t under prefix. defaults, when valid, is
// a value of t whose fields are the keys' defaults, as a behavior declares
// them; otherwise a key's default comes from Defaults.
func walk(
	t reflect.Type,
	prefix string,
	defaults reflect.Value,
) []Key {
	var keys []Key
	for i := range t.NumField() {
		f := t.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ",")
		if name == "" || !f.IsExported() {
			continue
		}
		if prefix != "" {
			name = prefix + "." + name
		}

		var fieldDefault reflect.Value
		if defaults.IsValid() {
			fieldDefault = defaults.Field(i)
		}

		ft := f.Type
		if ft.Kind() == reflect.Struct {
			keys = append(keys, walk(ft, name, fieldDefault)...)

			continue
		}

		// A zero value is no default, but false is a default of a switch.
		var def any
		if fieldDefault.IsValid() &&
			(!fieldDefault.IsZero() || fieldDefault.Kind() == reflect.Bool) {
			def = fieldDefault.Interface()
		}
		keys = append(keys, Key{
			Name:     name,
			Choice:   f.Tag.Get("choice"),
			Env:      EnvName(name),
			Type:     typeName(ft),
			Default:  def,
			Validate: f.Tag.Get("validate"),
			Doc:      f.Tag.Get("doc"),
			Secret:   f.Tag.Get("mask") == "password",
		})
	}

	return keys
}

func typeName(
	t reflect.Type,
) string {
	switch t.Kind() {
	case reflect.Slice:
		return "list"
	case reflect.Map:
		return "map"
	case reflect.Int, reflect.Int64:
		return "int"
	case reflect.Float64:
		return "float"
	default:
		return t.Kind().String()
	}
}
