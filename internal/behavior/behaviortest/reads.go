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

package behaviortest

import (
	"context"

	"go.uber.org/mock/gomock"

	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
)

// ReadsThroughRunCmd makes m answer each RunReadCmd, the read that gohai
// runs, with m's own RunCmd: a test of a provider then states each command
// once, as RunCmd, whether gohai or the provider runs it.
func ReadsThroughRunCmd(
	m *hostmocks.MockExec,
) {
	m.EXPECT().RunReadCmd(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, args []string) ([]byte, error) {
			out, err := m.RunCmd(ctx, name, args)

			return []byte(out), err
		}).
		AnyTimes()
}
