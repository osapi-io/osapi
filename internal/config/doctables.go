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
	"strings"

	"gopkg.in/yaml.v3"
)

// The markers of a generated config table on a docs page: the keys under
// each prefix that the start marker names, which DocTables writes between
// the two. A page keeps its own text around them.
const (
	docTableStart = "<!-- osapi:config "
	docTableEnd   = "<!-- /osapi:config -->"
)

// ShippedFile is the config that osapi ships, from the repository's root. A
// key with no default can have its value there.
const ShippedFile = "configs/osapi.yaml"

// Shipped is each value of a config file, by its dotted key.
func Shipped(
	data []byte,
) (map[string]any, error) {
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("read %s: %w", ShippedFile, err)
	}
	out := map[string]any{}
	var walk func(prefix string, m map[string]any)
	walk = func(prefix string, m map[string]any) {
		for k, v := range m {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			if child, ok := v.(map[string]any); ok {
				walk(key, child)

				continue
			}
			out[key] = v
		}
	}
	walk("", root)

	return out, nil
}

// DocTables is page with each generated config table written again, from
// Keys and the shipped config. A page with no markers is as it is.
func DocTables(
	page string,
	shipped map[string]any,
) string {
	var sb strings.Builder
	for {
		i := strings.Index(page, docTableStart)
		if i < 0 {
			sb.WriteString(page)

			return sb.String()
		}
		head := page[i+len(docTableStart):]
		closing := strings.Index(head, "-->")
		end := strings.Index(head, docTableEnd)
		if closing < 0 || end < closing {
			sb.WriteString(page)

			return sb.String()
		}
		prefixes := strings.Fields(head[:closing])
		sb.WriteString(page[:i+len(docTableStart)+closing+len("-->")])
		// Blank lines around the table, as Prettier writes them.
		sb.WriteString("\n\n" + DocTable(prefixes, shipped) + "\n\n")
		page = head[end:]
	}
}

// DocTable is the keys under prefixes as a markdown table, its columns
// padded as Prettier pads them: each key, its type, its default, the
// environment variable that sets it and what it is for. A key with no
// default says where its value comes from.
func DocTable(
	prefixes []string,
	shipped map[string]any,
) string {
	rows := [][]string{{"Key", "Type", "Default", "Environment", "Description"}}
	for _, k := range Keys() {
		if !under(k.Name, prefixes) {
			continue
		}
		rows = append(rows, []string{
			"`" + k.Name + "`",
			k.Type,
			defaultText(k, shipped),
			"`" + k.Env + "`",
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

// defaultText is k's default; the value of the shipped config for a key
// with none; or none.
func defaultText(
	k Key,
	shipped map[string]any,
) string {
	if k.Default != nil && fmt.Sprint(k.Default) != "" {
		return fmt.Sprintf("`%v`", k.Default)
	}
	if v, ok := shipped[k.Name]; ok {
		return fmt.Sprintf("none; `%v` in `%s`", v, ShippedFile)
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
