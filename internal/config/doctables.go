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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/internal/guide"
)

// DocTables is page with each generated config table written again, from
// Keys: a table goes between <!-- osapi:config prefixes --> and
// <!-- /osapi:config -->. A page with no markers is as it is.
func DocTables(
	page string,
) string {
	return guide.Fill(page, "config", DocTable)
}

// DocTable is the keys under prefixes as a markdown table, its columns
// padded as Prettier pads them: each key, its type, its default, the
// environment variable that sets it and what it is for. A key with no
// default says why: it is a choice of the operator.
func DocTable(
	prefixes []string,
) string {
	rows := [][]string{{"Key", "Type", "Default", "Environment", "Description"}}
	for _, k := range append(Keys(), VaultKeys()...) {
		if !under(k.Name, prefixes) {
			continue
		}
		env := "none"
		if k.Env != "" {
			env = "`" + k.Env + "`"
		}
		rows = append(rows, []string{
			"`" + k.Name + "`",
			k.Type,
			defaultText(k),
			env,
			strings.ReplaceAll(k.Doc, "|", "\\|"),
		})
	}

	return table(rows)
}

// under reports whether key is a prefix, or a key under one.
func under(
	key string,
	prefixes []string,
) bool {
	for _, p := range prefixes {
		if key == p || strings.HasPrefix(key, p+".") {
			return true
		}
	}

	return false
}

// defaultText is k's default, a list or a block as JSON; or none, with why
// it is a choice of the operator.
func defaultText(
	k Key,
) string {
	if k.Default != nil {
		switch reflect.ValueOf(k.Default).Kind() {
		case reflect.Slice, reflect.Map, reflect.Struct:
			data, _ := json.Marshal(k.Default)

			return fmt.Sprintf("`%s`", data)
		default:
			return fmt.Sprintf("`%v`", k.Default)
		}
	}
	if k.Choice != "" {
		return "none: " + k.Choice
	}

	return "none"
}

// table is rows as a markdown table, the first row its header, each column
// as wide as its widest cell.
func table(
	rows [][]string,
) string {
	widths := make([]int, len(rows[0]))
	for _, r := range rows {
		for i, c := range r {
			widths[i] = max(widths[i], len(c))
		}
	}
	line := func(cells []string) string {
		parts := make([]string, len(cells))
		for i, c := range cells {
			parts[i] = c + strings.Repeat(" ", widths[i]-len(c))
		}

		return "| " + strings.Join(parts, " | ") + " |"
	}
	dashes := make([]string, len(widths))
	for i, w := range widths {
		dashes[i] = strings.Repeat("-", w)
	}
	out := make([]string, 0, len(rows)+1)
	out = append(out, line(rows[0]), line(dashes))
	for _, r := range rows[1:] {
		out = append(out, line(r))
	}

	return strings.Join(out, "\n")
}
