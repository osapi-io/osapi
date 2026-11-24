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

package config

import osapiconfig "github.com/osapi-io/osapi/internal/config"

// The sample outputs of the examples, which the docs page of each command
// shows as the CLI prints it.
var (
	samplePort = Setting{
		Name:     "controller.api.port",
		Value:    8080,
		Source:   "file",
		Env:      "OSAPI_CONTROLLER_API_PORT",
		Type:     "int",
		Default:  8080,
		Validate: "min=1,max=65535",
		Doc:      "Port the REST API server listens on.",
	}

	sampleList = ListOutput{
		File: "/etc/osapi/osapi.yaml",
		Items: []Setting{
			samplePort,
			{
				Name:   "controller.api.security.signing_keys",
				Value:  "********",
				Source: "env",
				Env:    "OSAPI_CONTROLLER_API_SECURITY_SIGNING_KEYS",
				Type:   "list",
				Secret: true,
			},
			{
				Name:    "nats.server.port",
				Value:   4222,
				Source:  "default",
				Env:     "OSAPI_NATS_SERVER_PORT",
				Type:    "int",
				Default: 4222,
			},
		},
	}

	sampleGet = GetOutput{File: "/etc/osapi/osapi.yaml", Setting: samplePort}

	sampleValidate = ValidateOutput{
		File:     "/etc/osapi/osapi.yaml",
		Valid:    true,
		Problems: []osapiconfig.Problem{},
	}
)
