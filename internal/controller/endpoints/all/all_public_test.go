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

package all_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all/gen"
)

type AllPublicTestSuite struct {
	suite.Suite
}

// TestValidate holds each registered group to the rules of a definition.
func (s *AllPublicTestSuite) TestValidate() {
	s.NotEmpty(endpoint.All())
	for _, g := range endpoint.All() {
		s.Run(g.Name, func() {
			s.Empty(endpoint.Validate(g))
		})
	}
}

// TestRefusals sends each input that a method's Fails declares through the
// routes of every group, as a client does, and expects a 400 with an error
// body. Validate makes each rule of an input have one, so this covers each
// rule of each endpoint.
func (s *AllPublicTestSuite) TestRefusals() {
	const key = "test-signing-key-for-the-refusals"
	// The token has each method's permission, so a new group's, which no
	// role has yet, does not answer 403 before the input is checked.
	var perms []string
	for _, g := range endpoint.All() {
		for _, m := range g.Methods {
			if p := g.Permission(m); p != "" {
				perms = append(perms, string(p))
			}
		}
	}
	token, err := authtoken.New(slog.Default()).Generate(key, []string{"admin"}, "test", perms)
	s.Require().NoError(err)

	// Each group gets an empty Deps: the API refuses each input before a
	// handler would use one.
	var deps []any
	for _, g := range endpoint.All() {
		if g.Deps != nil {
			deps = append(deps, reflect.New(g.Deps).Elem().Interface())
		}
	}
	e := echo.New()
	for _, mount := range server.Handler(slog.Default(), key, nil, deps...) {
		mount(e)
	}

	count := 0
	for _, g := range endpoint.All() {
		for _, m := range g.Methods {
			for _, r := range m.Refusals() {
				count++
				s.Run(g.Name+" "+m.Name+": "+r.Doc, func() {
					path, body, contentType, err := command.Build(g, m, r.Input)
					s.Require().NoError(err)
					req := httptest.NewRequest(m.HTTPMethod(), path, bytes.NewReader(body))
					req.Header.Set("Content-Type", contentType)
					req.Header.Set("Authorization", "Bearer "+token)
					rec := httptest.NewRecorder()
					e.ServeHTTP(rec, req)

					s.Equal(http.StatusBadRequest, rec.Code, rec.Body.String())
					var answer struct {
						Error string `json:"error"`
					}
					s.Require().
						NoError(json.Unmarshal(rec.Body.Bytes(), &answer), rec.Body.String())
					s.NotEmpty(answer.Error)
				})
			}
		}
	}
	s.NotZero(count)
}

// TestHandlersRunFromRoutes fails a call of a handler from code outside its
// group. The route checks the input against its validate tags before the
// handler runs, so a direct call would skip that check.
func (s *AllPublicTestSuite) TestHandlersRunFromRoutes() {
	handlers := map[string]string{}
	dirs := map[string]string{}
	for _, g := range endpoint.All() {
		for _, m := range g.Methods {
			if m.Handler() == nil {
				continue
			}
			name := runtime.FuncForPC(reflect.ValueOf(m.Handler()).Pointer()).Name()
			handlers[name] = g.Name + " " + m.Name
			pkg := name[:strings.LastIndex(name, ".")]
			// A group's own package, and its generated routes, which are
			// how a handler runs.
			dirs[strings.TrimPrefix(pkg, module+"/")] = pkg
			dirs[strings.TrimPrefix(pkg, module+"/")+"/gen"] = pkg
		}
	}
	s.Require().NotEmpty(handlers)

	s.Empty(directCalls(root, handlers, dirs))
}

func TestAllPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AllPublicTestSuite))
}

// module is the import path of the repository, and root its directory from
// this package.
const (
	module = "github.com/osapi-io/osapi"
	root   = "../../../.."
)

// directCalls is each call, in the Go files under root that are not tests
// and not in a group's own package, of a function that handlers names.
func directCalls(
	root string,
	handlers map[string]string,
	own map[string]string,
) []string {
	var problems []string
	fset := token.NewFileSet()
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && rel != "." || d.Name() == "node_modules" {
				return filepath.SkipDir
			}

			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		if _, ok := own[filepath.ToSlash(filepath.Dir(rel))]; ok {
			return nil
		}
		f, err := parser.ParseFile(fset, p, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil //nolint:nilerr // a file that does not parse is the compiler's to report
		}
		imports := map[string]string{}
		for _, imp := range f.Imports {
			ip, _ := strconv.Unquote(imp.Path.Value)
			local := path.Base(ip)
			if imp.Name != nil {
				local = imp.Name.Name
			}
			imports[local] = ip
		}
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			x, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if m, ok := handlers[imports[x.Name]+"."+sel.Sel.Name]; ok {
				problems = append(problems, fmt.Sprintf(
					"%s uses %s.%s, the handler of %s, outside its route: the route checks the input first",
					fset.Position(sel.Pos()),
					x.Name,
					sel.Sel.Name,
					m,
				))
			}

			return true
		})

		return nil
	})

	return problems
}
