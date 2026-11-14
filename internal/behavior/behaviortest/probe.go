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

package behaviortest

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/validation"
)

// probeMark is in every probe, so a probe that got through is found by it.
const probeMark = "osapi-probe"

// probes are the hostile values each string input field is given in turn:
// a path that climbs out of its directory, a value that starts a line of its
// own, and a value that reads as a command-line option.
var probes = []struct {
	name  string
	value func(original string) string
}{
	{name: "path traversal", value: func(string) string { return "../../../../" + probeMark }},
	{name: "a new line", value: func(v string) string { return v + "\nosapi_probe = 1" }},
	{name: "an option", value: func(string) string { return "--" + probeMark }},
}

// checkProbes gives each string field of ex's input each probe in turn, on
// a fresh host set up as ex says. A probe passes when the API's validate
// tags refuse it, when the method fails, or when it changes nothing it
// should not: no file outside the directories the example itself touches,
// no file holding the probe's line beside other content, no command given
// the probe as an option, and no shell script holding it. An internal
// behavior is not probed: its input comes from other behaviors, not from a
// request.
func checkProbes(
	b *behavior.Behavior,
	m *behavior.Method,
	platform string,
	ex behavior.Example,
) []string {
	in := reflect.ValueOf(ex.Input)
	if b.Internal || in.Kind() != reflect.Struct {
		return nil
	}

	h, p, problem := prepare(b, platform, ex)
	if problem != "" {
		return nil
	}
	before := files(h)
	_, _ = call(m, p, ex.Input)
	dirs := map[string]bool{}
	for path := range changed(before, files(h)) {
		dirs[filepath.Dir(path)] = true
	}

	var problems []string
	for i := range in.NumField() {
		f := in.Type().Field(i)
		if !f.IsExported() || f.Type.Kind() != reflect.String {
			continue
		}
		for _, pr := range probes {
			probed := reflect.New(in.Type()).Elem()
			probed.Set(in)
			value := pr.value(in.Field(i).String())
			probed.Field(i).SetString(value)

			if msg := probe(b, m, platform, ex, probed.Interface(), value, dirs); msg != "" {
				problems = append(problems,
					fmt.Sprintf("with %s in %s, %q: %s", pr.name, f.Name, value, msg))
			}
		}
	}

	return problems
}

// probe runs the method with input on a fresh host and reports what the
// probe value got into, or "".
func probe(
	b *behavior.Behavior,
	m *behavior.Method,
	platform string,
	ex behavior.Example,
	input any,
	value string,
	dirs map[string]bool,
) string {
	if _, ok := validation.Struct(input); !ok {
		return ""
	}

	h, p, _ := prepare(b, platform, ex)
	before := files(h)
	if _, err := call(m, p, input); err != nil {
		return ""
	}

	for path, content := range changed(before, files(h)) {
		if strings.Contains(path, probeMark) && !dirs[filepath.Dir(path)] {
			return "it wrote " + path + ", outside the directories the example writes"
		}
		if strings.Contains(content, "\nosapi_probe = 1") && content != value {
			return "its line reached " + path
		}
	}
	for _, line := range h.ran {
		args := strings.Fields(line)
		if (args[0] == "sh" || args[0] == "bash") &&
			slices.Contains(args, "-c") && strings.Contains(line, probeMark) {
			return "it reached a shell script: " + line
		}
		for j, arg := range args {
			if arg == "--"+probeMark && !slices.Contains(args[:j], "--") {
				return "it reached a command as an option: " + line
			}
		}
	}

	return ""
}

// files is every file on h, by path, with its content.
func files(
	h *fakeHost,
) map[string]string {
	out := map[string]string{}
	_ = h.FS.WalkDir("/", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			data, _ := h.FS.ReadFile(path)
			out[path] = string(data)
		}

		return nil
	})

	return out
}

// changed is every file in after that is new or different from before.
func changed(
	before map[string]string,
	after map[string]string,
) map[string]string {
	out := map[string]string{}
	for path, content := range after {
		if old, ok := before[path]; !ok || old != content {
			out[path] = content
		}
	}
	return out
}
