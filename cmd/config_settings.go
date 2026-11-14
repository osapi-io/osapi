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

package cmd

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/config"
)

// setting is one key with its value in effect: everything config list and
// config get print, and what their --json carries.
type setting struct {
	config.Key
	Value  any    `json:"value"`
	Source string `json:"source"`
}

// redacted stands in for a secret's value.
const redacted = "********"

// readSettings reads the config file and the environment the way osapi start
// does, without checking them, and returns each key with its value and where
// it came from. A missing file is not an error: file says so, and the values
// are the defaults and the environment.
func readSettings() (file string, settings []setting, err error) {
	file = viper.GetString("osapiFile")
	v := viper.New()
	var c config.Config
	found, err := config.Read(v, file, &c)
	if err != nil {
		return file, nil, err
	}
	if !found {
		file += " (not found: defaults and environment only)"
	}

	for _, k := range config.Keys() {
		val := config.Value(v, k)
		if k.Secret && val != nil && val != "" {
			val = redacted
		}
		settings = append(settings, setting{Key: k, Value: val, Source: config.Source(v, k.Name)})
	}

	return file, settings, nil
}

// display is a value as text: empty for nothing.
func display(
	v any,
) string {
	if v == nil {
		return ""
	}

	return fmt.Sprint(v)
}
