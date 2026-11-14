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

package behavior_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type RulePublicTestSuite struct {
	suite.Suite
}

type ruled struct {
	Code string `json:"code" validate:"rule_test_code"`
}

func (s *RulePublicTestSuite) TestValidator() {
	var seen []string
	behavior.Define("early", "x",
		behavior.Validator("rule_test_early", func(string) bool { return true },
			behavior.Pattern("^a$"), "is wrong"),
	)

	behavior.OnRule(func(r behavior.Rule) { seen = append(seen, r.Tag) })
	s.Contains(seen, "rule_test_early", "a rule declared before OnRule is replayed")

	behavior.Define("late", "x",
		behavior.Validator("rule_test_code", func(v string) bool { return v == "ok" },
			behavior.Pattern("^ok$"), "is not ok"),
	)
	s.Contains(seen, "rule_test_code", "a rule declared after OnRule is passed on")

	schema, err := behavior.SchemaFor(reflect.TypeFor[ruled]())
	s.Require().NoError(err)
	s.Equal("^ok$", schema.Properties["code"].Pattern, "the spec carries the rule's schema")
}

func TestRulePublicTestSuite(t *testing.T) {
	suite.Run(t, new(RulePublicTestSuite))
}
