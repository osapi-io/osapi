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
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// FlagName is the flag that sets key: the key with its dots and
// underscores as dashes, such as --agent-max-jobs for agent.max_jobs, as
// OSAPI_AGENT_MAX_JOBS is its environment variable.
func FlagName(
	key string,
) string {
	return strings.NewReplacer(".", "-", "_", "-").Replace(key)
}

// BindFlags adds a flag to fs for each of osapi's keys, named by FlagName,
// with the key's type, default and doc, and binds it to the key in viper.
// The generated command of a process calls it for each key its Flag names.
// A key that is not osapi's is a mistake in the code, so it panics.
func BindFlags(
	fs *pflag.FlagSet,
	keys ...string,
) {
	own := map[string]Key{}
	for _, k := range walk(reflect.TypeFor[Config](), "", reflect.Value{}) {
		own[k.Name] = k
	}
	for _, name := range keys {
		k, ok := own[name]
		if !ok {
			panic(fmt.Sprintf("config: no key %q for a flag", name))
		}
		AddFlag(fs, viper.GetViper(), k, FlagName(name))
	}
}

// AddFlag adds the flag name to fs for k, with its type, default and doc,
// and binds it to k in v: v resolves the flag, then the environment, then
// the file, then the default, when the config is read.
func AddFlag(
	fs *pflag.FlagSet,
	v *viper.Viper,
	k Key,
	name string,
) {
	usage := fmt.Sprintf("%s (%s)", strings.TrimSuffix(k.Doc, "."), k.Name)
	switch k.Type {
	case "int":
		d, _ := k.Default.(int)
		fs.Int(name, d, usage)
	case "bool":
		d, _ := k.Default.(bool)
		fs.Bool(name, d, usage)
	case "float":
		d, _ := k.Default.(float64)
		fs.Float64(name, d, usage)
	case "list":
		d, _ := k.Default.([]string)
		fs.StringSlice(name, d, usage)
	default:
		d, _ := k.Default.(string)
		fs.String(name, d, usage)
	}
	_ = v.BindPFlag(k.Name, fs.Lookup(name))
}
