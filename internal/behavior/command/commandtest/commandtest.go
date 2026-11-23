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

// Package commandtest runs a behavior's generated commands for the
// gen/cli_public_test.go each one gets, with the SDK call faked.
package commandtest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	endpointcommand "github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Request is what a command sent to the API.
type Request struct {
	Method string
	Path   string
	Body   string
}

// caller records the request a command sends, and fails it, so the command
// reports the error and prints nothing.
type caller struct {
	sent *Request
}

func (c caller) Send(
	_ context.Context,
	method string,
	path string,
	_ string,
	body []byte,
) (*client.Response[json.RawMessage], error) {
	*c.sent = Request{Method: method, Path: path, Body: string(body)}

	return nil, errors.New("not sent")
}

// Sent runs the command line args under osapi client, such as node sysctl
// update --target server1 --key vm.swappiness, through the command
// layer and the endpoint layer's runner, and returns the request the
// command sent to the API.
func Sent(
	t *testing.T,
	args []string,
) Request {
	var sent Request
	root := &cobra.Command{Use: "client"}
	endpointcommand.Use(func() endpointcommand.Caller { return caller{sent: &sent} })
	require.NoError(t, clicommand.Attach(root, clicommand.Deps{
		// The call fails, so the command reports the error and never asks
		// whether to print JSON.
		JSON:  testing.Short,
		Error: func(error) {},
	}))
	reset(root)
	root.SetArgs(args)
	require.NoError(t, root.Execute())

	return sent
}

// reset puts each flag of cmd and the commands under it back to its
// default. The commands are package values that each Sent attaches again,
// so a flag keeps what an earlier command line set: a list flag would keep
// the values of the example before.
func reset(
	cmd *cobra.Command,
) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed {
			return
		}
		if list, ok := f.Value.(pflag.SliceValue); ok {
			_ = list.Replace(nil)
		} else if f.Value.Type() == "stringToString" {
			// A map flag adds each Set to what it holds, and has no Replace:
			// give the flag a new, empty value.
			fresh := pflag.NewFlagSet(f.Name, pflag.ContinueOnError)
			fresh.StringToString(f.Name, nil, "")
			f.Value = fresh.Lookup(f.Name).Value
		} else {
			_ = f.Value.Set(f.DefValue)
		}
		f.Changed = false
	})
	for _, c := range cmd.Commands() {
		reset(c)
	}
}
