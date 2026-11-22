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

package collectors_test

import (
	"testing"

	"github.com/osapi-io/gohai/pkg/gohai"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior/collectors"
	"github.com/osapi-io/osapi/internal/guide"
)

type CollectorsPublicTestSuite struct {
	suite.Suite
}

func (s *CollectorsPublicTestSuite) TestAppend() {
	base := guide.Spec{
		Kinds: []guide.Kind{{Kind: "definition", Title: "Definition"}},
		Terms: []guide.Term{{Kind: "definition", Name: "Define"}},
	}
	got := collectors.Append(base)
	s.Len(base.Kinds, 1, "the spec it is given does not change")
	s.Equal(guide.Kind{Kind: collectors.Kind, Title: "gohai collectors"}, got.Kinds[1])
	s.Len(got.Of(collectors.Kind), len(gohai.Collectors()), "a term for each collector")
}

func (s *CollectorsPublicTestSuite) TestTerms() {
	tests := []struct {
		name     string
		spec     gohai.CollectorSpec
		wantDoes string
		wantWhen string
	}{
		{
			name: "when the collector reads files and runs commands",
			spec: gohai.CollectorSpec{
				Name: "dns", Package: "github.com/osapi-io/gohai/pkg/gohai/collectors/dns",
				Constructor: "dns.NewWith(src)",
				Fields:      []gohai.Field{{Name: "nameservers"}, {Name: "search"}},
				Reads: []gohai.Reads{
					{
						Platform: "debian",
						Files:    []string{"/etc/resolv.conf"},
						Commands: []string{"resolvectl dns", "resolvectl domain"},
					},
					{Platform: "darwin"},
				},
			},
			wantDoes: "Returns nameservers, search. Build it with dns.NewWith(h.Sources()).",
			wantWhen: "An example's fake host gives, on debian: Exists for /etc/resolv.conf, " +
				"Prints for resolvectl dns; resolvectl domain. On darwin: nothing. With content, " +
				"it reads more under those paths. gohai collector get --name dns shows it in full; " +
				"dns is its package.",
		},
		{
			name: "when the collector reads the network and has no fields",
			spec: gohai.CollectorSpec{
				Name:        "ec2",
				Constructor: "ec2.New()",
				Network:     "the metadata service of ec2, over HTTP",
			},
			wantDoes: "Returns no fields. Build it with ec2.New(); it reads the metadata service of ec2, over HTTP, not the host.",
			wantWhen: "Not on a host: an example cannot give it.",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := collectors.Terms([]gohai.CollectorSpec{tc.spec})
			s.Require().Len(got, 1)
			s.Equal(collectors.Kind, got[0].Kind)
			s.Equal(tc.spec.Name, got[0].Name)
			s.Equal(tc.wantDoes, got[0].Does)
			s.Equal(tc.wantWhen, got[0].When)
		})
	}
}

func TestCollectorsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CollectorsPublicTestSuite))
}
