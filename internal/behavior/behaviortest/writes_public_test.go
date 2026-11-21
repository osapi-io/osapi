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

package behaviortest_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// files is a provider that changes /etc/w/x one way per op.
type files struct{ h host.Host }

type filesIn struct {
	Op string `json:"op"`
}

type filesOut struct {
	Changed bool `json:"changed"`
}

const filesPath = "/etc/w/x"

func changeFile(
	ctx context.Context,
	p files,
	in filesIn,
) (filesOut, error) {
	fs := p.h.FS
	_ = fs.MkdirAll("/etc/w", 0o755)
	var err error
	switch in.Op {
	case "open":
		var f interface{ Close() error }
		f, err = fs.OpenFile(filesPath, os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			_ = f.Close()
		}
	case "rename":
		// /tmp-x is the example's own file, so only the rename is a write.
		err = fs.Rename("/tmp-x", filesPath)
	case "remove":
		err = fs.Remove(filesPath)
	case "chmod":
		err = fs.Chmod(filesPath, 0o600)
	case "truncate":
		err = fs.Truncate(filesPath, 0)
	case "read":
		_, err = fs.ReadFile(filesPath)
	case "atomic":
		err = atomicfile.Write(ctx, fs, filesPath, []byte("x"), 0o644)
	default:
		err = fs.WriteFile(filesPath, []byte("x"), 0o644)
	}

	return filesOut{}, err
}

type WritesPublicTestSuite struct {
	suite.Suite
}

func (s *WritesPublicTestSuite) TestWrites() {
	tests := []struct {
		name   string
		op     string
		writes []string
		want   []string
	}{
		{
			name:   "a write under a declared pattern is fine",
			op:     "write",
			writes: []string{"/etc/w/*"},
		},
		{name: "any path is fine with *", op: "chmod", writes: []string{"*"}},
		{name: "a read is not a write", op: "read"},
		{name: "a write", op: "write", want: []string{"writes /etc/w/x"}},
		{name: "an open for writing", op: "open", want: []string{"writes /etc/w/x"}},
		{name: "a rename onto the path", op: "rename", want: []string{"writes /etc/w/x"}},
		{name: "a removal", op: "remove", want: []string{"writes /etc/w/x"}},
		{name: "a mode change", op: "chmod", want: []string{"writes /etc/w/x"}},
		{name: "a truncation", op: "truncate", want: []string{"writes /etc/w/x"}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("files", "Files", behavior.Node,
				behavior.Platform("debian", func(h host.Host) (files, error) {
					return files{h: h}, nil
				}),
				behavior.Writes(tc.writes...),
				behavior.Read("get", changeFile).Describe("Change it").
					Example("A change", filesIn{Op: tc.op}, filesOut{},
						// The example's own file is set up, not written.
						behavior.Exists(filesPath, "x"),
						behavior.Exists("/tmp-x", "x")),
			)

			var undeclared []string
			for _, p := range behaviortest.Check(b) {
				if strings.Contains(p, "does not declare it with behavior.Writes") {
					undeclared = append(undeclared, p)
				}
			}

			if tc.want == nil {
				s.Empty(undeclared)
				return
			}
			s.Require().NotEmpty(undeclared)
			s.Contains(undeclared[0], tc.want[0])
		})
	}
}

func (s *WritesPublicTestSuite) TestFileSpans() {
	tests := []struct {
		name string
		op   string
		want bool
	}{
		{name: "when a write goes through atomicfile it has its span", op: "atomic"},
		{name: "when a write goes through h.FS it has none", op: "write", want: true},
		{name: "when a removal goes through h.FS it has none", op: "remove", want: true},
		{name: "when a read goes through h.FS it needs none", op: "read"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("files", "Files", behavior.Node,
				behavior.Platform("debian", func(h host.Host) (files, error) {
					return files{h: h}, nil
				}),
				behavior.Writes("*"),
				behavior.Read("get", changeFile).Describe("Change it").
					Example("A change", filesIn{Op: tc.op}, filesOut{},
						behavior.Exists(filesPath, "x")),
			)

			var spanless []string
			for _, p := range behaviortest.Check(b) {
				if strings.Contains(p, "so the trace has no file span") {
					spanless = append(spanless, p)
				}
			}

			if !tc.want {
				s.Empty(spanless)
				return
			}
			s.Require().Len(spanless, 1)
			s.Contains(spanless[0], "changes /etc/w/x through h.FS")
			s.Contains(spanless[0], "Fix: For a file that osapi manages, require managedfile")
		})
	}
}

func (s *WritesPublicTestSuite) TestDirectCalls() {
	tools := behavior.Define("tools", "Tools", behavior.Node, behavior.Internal,
		behavior.Write("set", setNote))
	b := behavior.Define("user", "A user of tools", behavior.Node, behavior.Requires(tools))
	method := tools.Methods[0].Func()
	pkg := method[:strings.LastIndex(method, ".")]
	fn := method[strings.LastIndex(method, ".")+1:]

	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "a call through behavior.Call is fine",
			src: "package user\n\nimport (\n\t\"context\"\n\n\tt \"" + pkg + "\"\n" +
				"\t\"github.com/osapi-io/osapi/internal/behavior\"\n)\n\n" +
				"func f() { _, _ = behavior.Call(context.Background(), nil, t." + fn + ", nil, nil) }\n" +
				// A call of a local function, and of a field, are not of a package.
				"func h() {}\n" +
				"func g() { h(); var x struct{ y struct{ z func() } }; x.y.z() }\n",
		},
		{
			name: "a direct call is reported with its place and the fix",
			src: "package user\n\nimport t \"" + pkg + "\"\n\n" +
				"func f() { _, _ = t." + fn + "(nil, nil, nil) }\n",
			want: []string{"user.go:5:", "calls t." + fn + " directly", "behavior.Call"},
		},
		{
			name: "a file that does not parse is reported",
			src:  "package user\n\nfunc {",
			want: []string{"read "},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			dir := s.T().TempDir()
			s.Require().NoError(os.WriteFile(filepath.Join(dir, "user.go"), []byte(tc.src), 0o644))
			s.Require().NoError(os.WriteFile(filepath.Join(dir, "user_test.go"),
				[]byte("package user\n\nfunc {"), 0o644))
			defer behaviortest.SetSourceDir(dir)()

			r := &recorder{}
			behaviortest.Run(r, b)
			// user has no methods, so no example calls tools: that is the
			// used rule, which chain_public_test.go covers.
			var errs []string
			for _, e := range r.errs {
				if !strings.Contains(e, "no example calls it") {
					errs = append(errs, e)
				}
			}

			if tc.want == nil {
				s.Empty(errs)
				return
			}
			s.Require().Len(errs, 1)
			for _, w := range tc.want {
				s.Contains(errs[0], w)
			}
		})
	}

	s.Run("a behavior that requires nothing is not read", func() {
		defer behaviortest.SetSourceDir("/nonexistent")()

		r := &recorder{}
		behaviortest.Run(r, behavior.Define("alone", "Alone", behavior.Node))
		s.Empty(r.errs)
	})
}

func TestWritesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(WritesPublicTestSuite))
}
