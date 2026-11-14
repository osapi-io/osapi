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

package behavior

import (
	"reflect"
	"strings"
)

// SettingsPrefix is where every behavior's settings sit in osapi's config:
// behaviors.<name>.<field>.
const SettingsPrefix = "behaviors"

// Config declares a behavior's own settings. defaults is a struct whose
// fields carry mapstructure, doc and validate tags, the same as osapi's own
// config, and whose field values are the defaults. The settings join osapi's
// config under behaviors.<name>: config list shows them, OSAPI_ environment
// variables and agent start flags set them, and config validate checks them.
// Settings are standing configuration of the agent, which runs the methods;
// a method's per-call values are its input, never a setting.
func Config(
	defaults any,
) Option {
	return optionFunc(func(b *Behavior) { b.Defaults = defaults })
}

// SettingsKey is the config key the behavior's settings sit under, such as
// behaviors.sysctl.
func (b *Behavior) SettingsKey() string {
	return SettingsPrefix + "." + b.Name
}

// SettingsType is the type of the behavior's settings struct, or nil when it
// declares none.
func (b *Behavior) SettingsType() reflect.Type {
	if b.Defaults == nil {
		return nil
	}

	return reflect.TypeOf(b.Defaults)
}

// SetSettings stores the decoded settings. The config loader calls it once
// the config is read; a behavior reads them with Settings.
func (b *Behavior) SetSettings(
	v any,
) {
	b.settings = v
}

// CurrentSettings is b's decoded settings, or its declared defaults before
// the config is read, or nil when it declares none: what the agent puts in the
// host it builds b's provider with.
func (b *Behavior) CurrentSettings() any {
	if b.settings != nil {
		return b.settings
	}

	return b.Defaults
}

// Settings returns b's decoded settings as T, the struct passed to Config.
// Before the config is loaded, or in a process that never loads it, it
// returns the declared defaults.
func Settings[T any](
	b *Behavior,
) T {
	v, _ := b.CurrentSettings().(T)

	return v
}

// SettingFlag is the agent start flag for one setting, such as
// sysctl-dir: the behavior's name, then the key's path with dots and
// underscores as dashes.
func (b *Behavior) SettingFlag(
	path string,
) string {
	return b.Name + "-" + strings.NewReplacer(".", "-", "_", "-").Replace(path)
}
