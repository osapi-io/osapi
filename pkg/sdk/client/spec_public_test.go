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
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type SpecPublicTestSuite struct {
	suite.Suite
}

// TestServices: each service that Services lists is a field of Client, and
// each of its methods is a method of that field, so the description and
// the client cannot disagree.
func (s *SpecPublicTestSuite) TestServices() {
	services := client.Services()
	s.NotEmpty(services)
	c := reflect.TypeFor[client.Client]()
	for _, svc := range services {
		f, ok := c.FieldByName(svc.Name)
		s.Require().True(ok, "Client has the service %s", svc.Name)
		s.NotEmpty(svc.Doc, "%s says what it does", svc.Name)
		s.NotEmpty(svc.Methods, "%s has methods", svc.Name)
		for _, m := range svc.Methods {
			_, ok := f.Type.MethodByName(m.Name)
			s.True(ok, "%s has the method %s", svc.Name, m.Name)
			s.NotEmpty(m.HTTP, "%s.%s names its route", svc.Name, m.Name)
		}
	}
}

func TestSpecPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SpecPublicTestSuite))
}
