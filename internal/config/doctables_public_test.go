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
package config_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/config"
)

type DocTablesPublicTestSuite struct {
	suite.Suite
}

func TestDocTablesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DocTablesPublicTestSuite))
}

func (s *DocTablesPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *DocTablesPublicTestSuite) TestDocTables() {
	registerKV()
	page := "# Metrics\n\nBefore.\n\n" +
		"<!-- osapi:config controller.metrics.port agent.metrics.port -->\nold table\n<!-- /osapi:config -->\n\n" +
		"Between.\n\n<!-- osapi:config behaviors.kv.dir -->\n<!-- /osapi:config -->\n\nAfter.\n"

	got := config.DocTables(page)

	s.Equal("# Metrics\n\nBefore.\n\n"+
		"<!-- osapi:config controller.metrics.port agent.metrics.port -->\n\n"+
		"| Key                       | Type | Default | Environment                     | Description                         |\n"+
		"| ------------------------- | ---- | ------- | ------------------------------- | ----------------------------------- |\n"+
		"| `controller.metrics.port` | int  | `9090`  | `OSAPI_CONTROLLER_METRICS_PORT` | Port the metrics server listens on. |\n"+
		"| `agent.metrics.port`      | int  | `9091`  | `OSAPI_AGENT_METRICS_PORT`      | Port the metrics server listens on. |\n\n"+
		"<!-- /osapi:config -->\n\n"+
		"Between.\n\n<!-- osapi:config behaviors.kv.dir -->\n\n"+
		"| Key                | Type   | Default   | Environment              | Description      |\n"+
		"| ------------------ | ------ | --------- | ------------------------ | ---------------- |\n"+
		"| `behaviors.kv.dir` | string | `/var/kv` | `OSAPI_BEHAVIORS_KV_DIR` | Where it writes. |\n\n"+
		"<!-- /osapi:config -->\n\nAfter.\n",
		got,
		"each table is written again from the keys, and the text around it stays")
	s.Equal(got, config.DocTables(got), "a page written again does not change")

	s.Contains(
		config.DocTable([]string{"agent.hostname"}),
		"| none: empty is the host's own hostname ",
		"a key with no default says why it is a choice",
	)
	s.Contains(config.DocTable([]string{"behaviors.kv.dir"}), "Where it writes.")
	s.Regexp("\\| `controller\\.secrets\\.vaults\\[\\]\\.key_file` \\| string \\| none +\\| none ",
		config.DocTable([]string{"controller.secrets"}),
		"a setting of a vault has no environment variable")
	s.Contains(config.DocTable([]string{"controller.secrets"}),
		"/etc/osapi/secrets.key",
		"the default vault reads as JSON")

	for _, broken := range []string{
		"<!-- osapi:config agent.hostname",
		"<!-- osapi:config agent.hostname -->\nno end",
		"<!-- /osapi:config -->\n<!-- osapi:config agent.hostname -->",
	} {
		s.Equal(broken, config.DocTables(broken), "a marker with no end is left as it is")
	}
}
