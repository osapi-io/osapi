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
package instruments_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/telemetry/instruments"
)

type InstrumentsPublicTestSuite struct {
	suite.Suite
}

func TestInstrumentsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(InstrumentsPublicTestSuite))
}

func (s *InstrumentsPublicTestSuite) TestExported() {
	s.Equal("osapi_jobs_created_total", instruments.JobsCreated.Exported())
}

func (s *InstrumentsPublicTestSuite) TestAll() {
	seen := map[string]bool{}
	for _, i := range instruments.All() {
		s.False(seen[i.Name], "%s is listed once", i.Name)
		seen[i.Name] = true
		s.False(strings.HasPrefix(i.Name, instruments.Namespace+"_"),
			"%s: the exporter adds the namespace", i.Name)
		s.Contains(
			[]string{instruments.KindCounter, instruments.KindGauge, instruments.KindHistogram},
			i.Kind,
		)
		s.Contains([]string{"all", "agent", "controller"}, i.Component)
		s.NotEmpty(i.Doc)
	}
}

func (s *InstrumentsPublicTestSuite) TestTable() {
	table := instruments.Table()

	s.True(strings.HasPrefix(table, "| Metric | Type | Labels | Component | Description |\n"))
	s.Contains(
		table,
		"| `osapi_jobs_processed_total` | counter | status | agent | Total jobs processed. |",
	)
	s.Contains(table, "| `osapi_component_up` | gauge |  | all |", "no labels is an empty cell")
	s.Equal(len(instruments.All())+2, strings.Count(table, "\n")+1, "a row for each instrument")
}
