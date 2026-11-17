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
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/apicontract"
	"github.com/osapi-io/osapi/internal/endpoint"
	_ "github.com/osapi-io/osapi/internal/endpoints/all"
)

// repoRoot is the repository root, from this package's directory.
const repoRoot = "../../.."

// moving is each group whose hand-written spec the controller still serves,
// and that spec. A row goes when its group moves and the spec is deleted;
// when the table is empty, the test goes too.
var moving = []struct {
	group string
	spec  string
	// accepted are the lines that differ on purpose, each with its reason:
	// a line only the hand-written spec has starts with -, one only the
	// generated spec has with +.
	accepted map[string]string
}{
	{
		group: "agent",
		spec:  "internal/controller/api/agent/gen/api.yaml",
	},
}

type ParityPublicTestSuite struct {
	suite.Suite
}

// TestEndpointParity holds the generated spec of each group that moves to
// its hand-written spec: the same operations, parameters, auth, bodies,
// responses and schemas. Prose is not compared; osapi endpoint get --name
// <group> --openapi prints the generated spec to read beside the old one.
func (s *ParityPublicTestSuite) TestEndpointParity() {
	for _, m := range moving {
		s.Run("when "+m.group+" replaces "+m.spec, func() {
			g := endpoint.Lookup(m.group)
			s.Require().NotNil(g, "group %s is not registered", m.group)

			data, err := os.ReadFile(filepath.Join(repoRoot, m.spec))
			s.Require().NoError(err)
			var hand map[string]any
			s.Require().NoError(yaml.Unmarshal(data, &hand))

			doc, err := endpoint.OpenAPI(g, endpoint.ErrorRef)
			s.Require().NoError(err)
			raw, err := json.Marshal(doc)
			s.Require().NoError(err)
			var gen map[string]any
			s.Require().NoError(json.Unmarshal(raw, &gen))

			want := set(apicontract.Lines(hand))
			got := set(apicontract.Lines(gen))
			var diff []string
			for l := range want {
				if !got[l] {
					diff = append(diff, "-"+l)
				}
			}
			for l := range got {
				if !want[l] {
					diff = append(diff, "+"+l)
				}
			}
			var unexpected []string
			for _, d := range diff {
				if _, ok := m.accepted[d]; !ok {
					unexpected = append(unexpected, d)
				}
			}
			sort.Strings(unexpected)
			s.Empty(unexpected, "the generated spec of %s differs from %s:\n%s",
				m.group, m.spec, strings.Join(unexpected, "\n"))
		})
	}
}

func set(
	lines []string,
) map[string]bool {
	out := make(map[string]bool, len(lines))
	for _, l := range lines {
		out[l] = true
	}

	return out
}

func TestParityPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ParityPublicTestSuite))
}
