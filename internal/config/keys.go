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

	"github.com/osapi-io/osapi/internal/behavior"
)

// EnvPrefix is the prefix of every environment variable that sets a key.
const EnvPrefix = "OSAPI"

// Defaults are the values a key takes when neither the config file nor the
// environment sets it. Every name here is a key of Config, which a test
// checks, so a default cannot be set on a key that does not exist.
var Defaults = map[string]any{
	"controller.client.url":      "http://0.0.0.0:8080",
	"controller.api.port":        8080,
	"controller.api.job_timeout": "30s",
	"controller.metrics.port":    9090,
	"controller.nats.host":       "localhost",
	"controller.nats.port":       4222,

	"controller.pki.enabled":               false,
	"controller.pki.key_dir":               "/etc/osapi/pki",
	"controller.pki.auto_accept":           false,
	"controller.pki.rotation_grace_period": "24h",

	"nats.server.host":         "localhost",
	"nats.server.port":         4222,
	"nats.stream.name":         "JOBS",
	"nats.stream.subjects":     "jobs.>",
	"nats.kv.bucket":           "job-queue",
	"nats.kv.response_bucket":  "job-responses",
	"nats.registry.bucket":     "agent-registry",
	"nats.registry.ttl":        "30s",
	"nats.enrollment.bucket":   "agent-enrollment",
	"nats.enrollment.storage":  "file",
	"nats.enrollment.replicas": 1,

	"agent.nats.host":                            "localhost",
	"agent.nats.port":                            4222,
	"agent.max_jobs":                             10,
	"agent.queue_group":                          "job-agents",
	"agent.facts.interval":                       "60s",
	"agent.conditions.memory_pressure_threshold": 90,
	"agent.conditions.high_load_multiplier":      2.0,
	"agent.conditions.disk_pressure_threshold":   90,
	"agent.consumer.max_deliver":                 5,
	"agent.consumer.ack_wait":                    "2m",
	"agent.consumer.max_ack_pending":             1000,
	"agent.consumer.replay_policy":               "instant",
	"agent.pki.enabled":                          false,
	"agent.pki.key_dir":                          "/etc/osapi/pki",
}

// Key is one setting osapi reads: its dotted name, the environment variable
// that sets it, its type, its default, its constraints and what it is for.
type Key struct {
	Name     string `json:"name"`
	Env      string `json:"env"`
	Type     string `json:"type"`
	Default  any    `json:"default,omitempty"`
	Validate string `json:"validate,omitempty"`
	Doc      string `json:"doc,omitempty"`
	// Secret is true for a value that is never printed: a password, a
	// signing key or a token.
	Secret bool `json:"secret,omitempty"`
	// Behavior names the behavior that declared the key, for a behavior
	// setting.
	Behavior string `json:"behavior,omitempty"`
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
// declare them, then each registered behavior's, under behaviors.<name>. A
// struct is walked into; a list, a map or a scalar is one key.
func Keys() []Key {
	keys := walk(reflect.TypeFor[Config](), "", reflect.Value{})
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

// defaultOf is key's default, from Defaults or a behavior's declared
// settings.
func defaultOf(
	key string,
) (any, bool) {
	if d, ok := Defaults[key]; ok {
		return d, true
	}
	for _, k := range Keys() {
		if k.Name == key && k.Behavior != "" {
			return k.Default, true
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
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			keys = append(keys, walk(ft, name, fieldDefault)...)

			continue
		}

		def := Defaults[name]
		if fieldDefault.IsValid() {
			def = fieldDefault.Interface()
		}
		keys = append(keys, Key{
			Name:     name,
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
