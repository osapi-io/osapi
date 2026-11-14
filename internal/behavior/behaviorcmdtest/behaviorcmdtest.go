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

// Package behaviorcmdtest runs a behavior's generated commands for the
// gen/cli_public_test.go each one gets, with the SDK call mocked.
package behaviorcmdtest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviorcmd"
	"github.com/osapi-io/osapi/internal/behaviorcmd/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Request is what a command sent to the API.
type Request struct {
	Method string
	Path   string
	Body   string
}

// Sent runs the command line args under osapi client next, such as
// sysctl update --target server1 --key vm.swappiness, and returns the request
// the command sent to the API.
func Sent(
	t *testing.T,
	args []string,
) Request {
	var sent Request
	caller := mocks.NewMockCaller(gomock.NewController(t))
	caller.EXPECT().Request(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			_ context.Context,
			method string,
			path string,
			body []byte,
		) (*client.Response[json.RawMessage], error) {
			sent = Request{Method: method, Path: path, Body: string(body)}

			return nil, errors.New("not sent")
		})

	root := &cobra.Command{Use: "next"}
	root.PersistentFlags().String("target", "", "")
	behaviorcmd.Attach(root, behaviorcmd.Deps{
		Caller: func() behaviorcmd.Caller { return caller },
		// The call fails, so the command reports the error and never asks
		// whether to print JSON.
		JSON:  testing.Short,
		Error: func(error) {},
	})
	root.SetArgs(args)
	require.NoError(t, root.Execute())

	return sent
}
