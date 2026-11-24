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

package client_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/secret"
)

type MarshalDataPublicTestSuite struct {
	suite.Suite
}

// TestMarshalData: the job carries each secret as the caller sent it, so
// the agent gives the provider the value and not [redacted], until the
// controller seals it (osapi #1247).
func (s *MarshalDataPublicTestSuite) TestMarshalData() {
	type createInput struct {
		Name string                  `json:"name"`
		Env  map[string]secret.Value `json:"env,omitempty"`
	}

	data, err := client.ExportMarshalData(createInput{
		Name: "app", Env: map[string]secret.Value{"TOKEN": "hunter2"},
	})

	s.Require().NoError(err)
	s.JSONEq(`{"name":"app","env":{"TOKEN":"hunter2"}}`, string(data))
}

func TestMarshalDataPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MarshalDataPublicTestSuite))
}
