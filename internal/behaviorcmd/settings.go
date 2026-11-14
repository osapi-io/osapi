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
package behaviorcmd

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/config"
)

// SettingFlags adds a flag to fs for every behavior setting, named by
// Key.Flag, such as --sysctl-dir, and binds it to the setting's key in v
// the way agent.go binds its own flags. Behaviors register in package init,
// before this runs, so every setting is here; v resolves the flag, then the
// environment, then the file, then the default when the config is read.
func SettingFlags(
	fs *pflag.FlagSet,
	v *viper.Viper,
) {
	for _, k := range config.Keys() {
		if k.Flag == "" {
			continue
		}
		usage := fmt.Sprintf("%s (%s)", k.Doc, k.Name)
		switch k.Type {
		case "int":
			d, _ := k.Default.(int)
			fs.Int(k.Flag, d, usage)
		case "bool":
			d, _ := k.Default.(bool)
			fs.Bool(k.Flag, d, usage)
		case "float":
			d, _ := k.Default.(float64)
			fs.Float64(k.Flag, d, usage)
		case "list":
			d, _ := k.Default.([]string)
			fs.StringSlice(k.Flag, d, usage)
		default:
			d, _ := k.Default.(string)
			fs.String(k.Flag, d, usage)
		}
		_ = v.BindPFlag(k.Name, fs.Lookup(k.Flag))
	}
}
