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

package facts_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/facts"
)

type KeysPublicTestSuite struct {
	suite.Suite
}

func (s *KeysPublicTestSuite) TestBuiltInKeys() {
	keys := facts.BuiltInKeys()

	s.Len(keys, 5)
	s.Contains(keys, facts.KeyInterfacePrimary)
	s.Contains(keys, facts.KeyHostname)
	s.Contains(keys, facts.KeyArch)
	s.Contains(keys, facts.KeyKernel)
	s.Contains(keys, facts.KeyFQDN)
}

func (s *KeysPublicTestSuite) TestBuiltInKeysReturnsNewSlice() {
	a := facts.BuiltInKeys()
	b := facts.BuiltInKeys()
	a[0] = "mutated"
	s.NotEqual(a[0], b[0], "BuiltInKeys should return a new slice each call")
}

func (s *KeysPublicTestSuite) TestIsKnownKey() {
	tests := []struct {
		name   string
		key    string
		expect bool
	}{
		{
			name:   "when interface.primary",
			key:    facts.KeyInterfacePrimary,
			expect: true,
		},
		{
			name:   "when hostname",
			key:    facts.KeyHostname,
			expect: true,
		},
		{
			name:   "when arch",
			key:    facts.KeyArch,
			expect: true,
		},
		{
			name:   "when kernel",
			key:    facts.KeyKernel,
			expect: true,
		},
		{
			name:   "when fqdn",
			key:    facts.KeyFQDN,
			expect: true,
		},
		{
			name:   "when valid custom key",
			key:    "custom.gateway",
			expect: true,
		},
		{
			name:   "when valid custom key with dots",
			key:    "custom.network.gateway",
			expect: true,
		},
		{
			name:   "when custom prefix only",
			key:    "custom.",
			expect: false,
		},
		{
			name:   "when empty string",
			key:    "",
			expect: false,
		},
		{
			name:   "when unknown key",
			key:    "unknown",
			expect: false,
		},
		{
			name:   "when partial match",
			key:    "host",
			expect: false,
		},
		{
			name:   "when not fact prefix",
			key:    "@notfact.x",
			expect: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.expect, facts.IsKnownKey(tc.key))
		})
	}
}

func (s *KeysPublicTestSuite) TestIsCustomKey() {
	tests := []struct {
		name   string
		key    string
		expect bool
	}{
		{
			name:   "when valid custom key",
			key:    "custom.gateway",
			expect: true,
		},
		{
			name:   "when valid custom key with nested dots",
			key:    "custom.network.primary.gateway",
			expect: true,
		},
		{
			name:   "when custom prefix only",
			key:    "custom.",
			expect: false,
		},
		{
			name:   "when empty string",
			key:    "",
			expect: false,
		},
		{
			name:   "when built-in key",
			key:    "hostname",
			expect: false,
		},
		{
			name:   "when partial custom prefix",
			key:    "custo",
			expect: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.expect, facts.IsCustomKey(tc.key))
		})
	}
}

func (s *KeysPublicTestSuite) TestConstants() {
	s.Equal("interface.primary", facts.KeyInterfacePrimary)
	s.Equal("hostname", facts.KeyHostname)
	s.Equal("arch", facts.KeyArch)
	s.Equal("kernel", facts.KeyKernel)
	s.Equal("fqdn", facts.KeyFQDN)
	s.Equal("custom.", facts.CustomPrefix)
	s.Equal("@fact.", facts.Prefix)
}

func TestKeysPublicTestSuite(t *testing.T) {
	suite.Run(t, new(KeysPublicTestSuite))
}
