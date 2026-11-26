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

package endpoint

import "reflect"

// SettingsPrefix is where each endpoint group's settings sit in osapi's
// config: controller.endpoints.<name>.<field>.
const SettingsPrefix = "controller.endpoints"

// Config declares a group's own settings, as behavior.Config does for a
// behavior. defaults is a struct whose fields carry mapstructure, doc and
// validate tags, the same as osapi's own config, and whose field values are
// the defaults. The settings join osapi's config under
// controller.endpoints.<name>: config list shows them, OSAPI_ environment
// variables set them, and config validate checks them. A handler reads them
// with Settings. Settings are standing configuration of the controller; a
// request's values are its input, never a setting.
func Config(
	defaults any,
) Option {
	return optionFunc(func(g *Group) { g.Defaults = defaults })
}

// SettingsKey is the config key the group's settings sit under, such as
// controller.endpoints.job.
func (g *Group) SettingsKey() string {
	return SettingsPrefix + "." + g.Name
}

// SettingsType is the type of the group's settings struct, or nil when it
// declares none.
func (g *Group) SettingsType() reflect.Type {
	if g.Defaults == nil {
		return nil
	}

	return reflect.TypeOf(g.Defaults)
}

// SetSettings stores the decoded settings. The config loader calls it once
// the config is read.
func (g *Group) SetSettings(
	v any,
) {
	g.settings = v
}

// CurrentSettings is g's decoded settings, or its declared defaults before
// the config is read, or nil when it declares none.
func (g *Group) CurrentSettings() any {
	if g.settings != nil {
		return g.settings
	}

	return g.Defaults
}

// Settings returns g's decoded settings as T, the struct passed to Config.
// Before the config is loaded, or in a process that never loads it, it
// returns the declared defaults.
func Settings[T any](
	g *Group,
) T {
	v, _ := g.CurrentSettings().(T)

	return v
}
