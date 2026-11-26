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

package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type EnumsPublicTestSuite struct {
	suite.Suite
}

// TestDoc: a type that describes itself in the spec says what it is.
func (s *EnumsPublicTestSuite) TestDoc() {
	s.Contains(types.ContentRaw.Doc(), "how a deploy treats a file")
}

// TestValid: a value of a closed set is valid when its type lists it.
func (s *EnumsPublicTestSuite) TestValid() {
	tests := []struct {
		name  string
		valid func() bool
		want  bool
	}{
		{name: "when a content type is listed", valid: types.ContentTemplate.Valid, want: true},
		{name: "when a content type is not listed", valid: types.ContentType("zip").Valid},
		{name: "when a reason is listed", valid: types.ReasonNotManaged.Valid, want: true},
		{name: "when a reason is not listed", valid: types.Reason("not_manged").Valid},
		{name: "when a permission is listed", valid: types.PermFileRead.Valid, want: true},
		{name: "when a permission is not listed", valid: types.Permission("helth:read").Valid},
		{name: "when a job status is listed", valid: types.JobStatusSkipped.Valid, want: true},
		{name: "when a job status is not listed", valid: types.JobStatus("done").Valid},
		{name: "when a method kind is listed", valid: types.KindAction.Valid, want: true},
		{name: "when a method kind is not listed", valid: types.MethodKind("delete").Valid},
		{name: "when an agent state is listed", valid: types.AgentDraining.Valid, want: true},
		{name: "when an agent state is not listed", valid: types.AgentState("Busy").Valid},
		{name: "when a condition is listed", valid: types.ConditionDrainStuck.Valid, want: true},
		{name: "when a condition is not listed", valid: types.ConditionType("Hot").Valid},
		{
			name:  "when an admission state is listed",
			valid: types.AdmissionRevoked.Valid,
			want:  true,
		},
		{
			name:  "when an admission state is not listed",
			valid: types.AdmissionState("accepted").Valid,
		},
		{name: "when a log level is listed", valid: types.LogWarn.Valid, want: true},
		{name: "when a log level is not listed", valid: types.LogLevel("verbose").Valid},
		{name: "when a Signal is listed", valid: types.SignalTerm.Valid, want: true},
		{name: "when a Signal is not listed", valid: types.Signal("nope").Valid},
		{name: "when a ContainerState is listed", valid: types.ContainerRunning.Valid, want: true},
		{name: "when a ContainerState is not listed", valid: types.ContainerState("nope").Valid},
		{name: "when a Interval is listed", valid: types.IntervalHourly.Valid, want: true},
		{name: "when a Interval is not listed", valid: types.Interval("nope").Valid},
		{name: "when a InterfaceType is listed", valid: types.InterfaceEthernet.Valid, want: true},
		{name: "when a InterfaceType is not listed", valid: types.InterfaceType("nope").Valid},
		{name: "when a AddressFamily is listed", valid: types.FamilyInet.Valid, want: true},
		{name: "when a AddressFamily is not listed", valid: types.AddressFamily("nope").Valid},
		{name: "when a Lock is listed", valid: types.Locked.Valid, want: true},
		{name: "when a Lock is not listed", valid: types.Lock("nope").Valid},
		{name: "when a SpanStatus is listed", valid: types.SpanOK.Valid, want: true},
		{name: "when a SpanStatus is not listed", valid: types.SpanStatus("nope").Valid},
		{
			name:  "when a ComponentType is listed",
			valid: types.ComponentController.Valid,
			want:  true,
		},
		{name: "when a ComponentType is not listed", valid: types.ComponentType("nope").Valid},
		{name: "when a FieldIn is listed", valid: types.FieldInPath.Valid, want: true},
		{name: "when a FieldIn is not listed", valid: types.FieldIn("nope").Valid},
		{name: "when a FieldKind is listed", valid: types.FieldString.Valid, want: true},
		{name: "when a FieldKind is not listed", valid: types.FieldKind("nope").Valid},
		{name: "when a ColumnKind is listed", valid: types.ColumnValue.Valid, want: true},
		{name: "when a ColumnKind is not listed", valid: types.ColumnKind("nope").Valid},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.valid())
		})
	}
}

func TestEnumsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(EnumsPublicTestSuite))
}

// TestAll: each closed set of the SDK is listed once, says what it is, and
// lists values that it holds valid.
func (s *EnumsPublicTestSuite) TestAll() {
	all := types.All()
	s.NotEmpty(all)
	seen := map[string]bool{}
	for _, t := range all {
		s.False(seen[t.Name], "%s is listed once", t.Name)
		seen[t.Name] = true
		s.NotEmpty(t.Doc, "%s says what it is", t.Name)
		s.NotEmpty(t.Values, "%s has values", t.Name)
	}
	s.True(seen["ContentType"])
	s.Contains(types.ContentRaw.Doc(), "how a deploy treats a file")
}
