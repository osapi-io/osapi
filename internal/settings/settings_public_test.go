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

package settings_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/settings"
)

type SettingsPublicTestSuite struct {
	suite.Suite
}

func (s *SettingsPublicTestSuite) TestOwn() {
	settings.Own("zzb", "zz b start", 2)
	settings.Own("zza", "zz a start", 1)

	got, ok := settings.Of("zza")
	s.Require().True(ok)
	s.Equal(settings.Owner{Prefix: "zza", By: "zz a start", Defaults: 1}, got)
	_, ok = settings.Of("zznone")
	s.False(ok)
	_, ok = settings.Of("zzb")
	s.True(ok)

	s.PanicsWithValue(`settings: zz a start and zz again both own "zza"`, func() {
		settings.Own("zza", "zz again", 3)
	})
}

func TestSettingsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SettingsPublicTestSuite))
}
