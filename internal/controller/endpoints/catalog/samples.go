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

package catalog

import (
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// sampleCatalog is the sample answer of the example, which the docs page of
// the command shows as the CLI prints it: one behavior with one method, and
// one of each other part.
var sampleCatalog = CatalogResponse{
	Behaviors: []CatalogBehavior{{
		Name:      "hostname",
		Title:     "Hostname",
		Area:      "node",
		Doc:       "The hostname of a host.",
		Platforms: []string{"debian", "darwin"},
		Uses:      []string{},
		UsedBy:    []string{},
		Methods: []CatalogMethod{{
			Name:       "get",
			Kind:       "read",
			Doc:        "Get the hostname.",
			HTTPMethod: "GET",
			Path:       "/api/node/{hostname}/hostname",
			Permission: types.PermHostnameRead,
			Platforms:  []string{"debian", "darwin"},
			RunsOn:     []string{"debian", "darwin"},
			Input:      "GetInput {}",
			Output:     "GetOutput { hostname }",
			Fields:     []CatalogField{},
			Columns: []CatalogColumn{
				{Name: "hostname", Label: "Hostname", Doc: "The hostname.", Kind: "value"},
			},
			Calls:    []CatalogCall{},
			Examples: []CatalogExample{},
		}},
	}},
	Areas:       []CatalogArea{{Name: "node", Label: "Node"}},
	Permissions: []types.Permission{types.PermHostnameRead, types.PermHostnameWrite},
	Roles: []CatalogRole{{
		Name:        "observer",
		Description: "A dashboard or a monitor.",
		Permissions: []types.Permission{types.PermHealthRead},
	}},
	Operations: []CatalogOperation{
		{ID: "GetNodeHostname", Permission: types.PermHostnameRead},
	},
	StatusTones:  map[string]string{"ok": "ok", "failed": "err"},
	CountTones:   map[string]string{"failed": "err"},
	HostFamilies: []string{"debian", "darwin"},
}
