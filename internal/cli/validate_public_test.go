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

package cli_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/cli"
)

type ValidatePublicTestSuite struct {
	suite.Suite
}

func TestValidatePublicTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatePublicTestSuite))
}

func (suite *ValidatePublicTestSuite) TestIsOSFamilySupported() {
	tests := []struct {
		name       string
		distro     string
		version    string
		wantFamily string
		wantOK     bool
	}{
		{
			name:       "when debian 12 is supported",
			distro:     "debian",
			version:    "12",
			wantFamily: "Debian",
			wantOK:     true,
		},
		{
			name:       "when debian 12 point release is supported",
			distro:     "debian",
			version:    "12.13",
			wantFamily: "Debian",
			wantOK:     true,
		},
		{
			name:       "when debian 13 is supported",
			distro:     "debian",
			version:    "13",
			wantFamily: "Debian",
			wantOK:     true,
		},
		{
			name:       "when ubuntu 20.04 is supported",
			distro:     "ubuntu",
			version:    "20.04",
			wantFamily: "Debian",
			wantOK:     true,
		},
		{
			name:       "when ubuntu 22.04 is supported",
			distro:     "ubuntu",
			version:    "22.04",
			wantFamily: "Debian",
			wantOK:     true,
		},
		{
			name:       "when ubuntu 24.04 is supported",
			distro:     "ubuntu",
			version:    "24.04",
			wantFamily: "Debian",
			wantOK:     true,
		},
		{
			name:       "when Ubuntu with uppercase is supported",
			distro:     "Ubuntu",
			version:    "24.04",
			wantFamily: "Debian",
			wantOK:     true,
		},
		{
			name:       "when unsupported distro returns false",
			distro:     "centos",
			version:    "8",
			wantFamily: "",
			wantOK:     false,
		},
		{
			name:       "when unsupported version returns false",
			distro:     "ubuntu",
			version:    "18.04",
			wantFamily: "",
			wantOK:     false,
		},
		{
			name:       "when empty distro returns false",
			distro:     "",
			version:    "24.04",
			wantFamily: "",
			wantOK:     false,
		},
		{
			name:       "when empty version returns false",
			distro:     "ubuntu",
			version:    "",
			wantFamily: "",
			wantOK:     false,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			family, ok := cli.IsOSFamilySupported(tc.distro, tc.version)

			suite.Equal(tc.wantFamily, family)
			suite.Equal(tc.wantOK, ok)
		})
	}
}
