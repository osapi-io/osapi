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

package natsperms

import (
	"context"
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/bus"
)

// Get is the subjects of the role, for the streams and buckets of the config
// under the namespace of controller.nats.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (GetOutput, error) {
	namespace := d.Config.Controller.NATS.Namespace
	p, err := bus.RolePermissions(bus.Role(in.Role), namespace, d.Config.NATS)
	if err != nil {
		return GetOutput{}, err
	}

	return GetOutput{
		Role:      in.Role,
		Namespace: namespace,
		Publish:   p.Publish,
		Subscribe: p.Subscribe,
		Responses: p.Responses,
		NSC:       nscFlags(p),
	}, nil
}

// nscFlags are the flags of nsc add user that give p.
func nscFlags(
	p bus.Permissions,
) string {
	flags := fmt.Sprintf("--allow-pub %q --allow-sub %q",
		strings.Join(p.Publish, ","), strings.Join(p.Subscribe, ","))
	if p.Responses {
		flags += " --allow-pub-response 1"
	}

	return flags
}
