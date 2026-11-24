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
	"maps"
	"slices"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// GetInput is the input of get, which has none.
type GetInput struct{}

// OperationID is the operationId of an operation of the API, such as
// GetAgents. Its schema is an enum of each operation that needs a
// permission.
type OperationID string

// Values are each operation that needs a permission, in order.
func (OperationID) Values() []OperationID {
	ops := authtoken.Operations()
	ids := make([]OperationID, 0, len(ops))
	for _, id := range slices.Sorted(maps.Keys(ops)) {
		ids = append(ids, OperationID(id))
	}

	return ids
}

// Valid reports whether id is an operation that needs a permission.
func (id OperationID) Valid() bool {
	_, ok := authtoken.Operations()[string(id)]

	return ok
}

// CatalogResponse is what the UI reads from the controller after sign-in:
// the behaviors, the areas, the permissions and the roles, the permission
// of each operation, the tones of the statuses and the counts, and the
// host families.
type CatalogResponse struct { //nolint:revive // the schema name is part of the API
	_            struct{}           `title:"Catalog"`
	Behaviors    []CatalogBehavior  `                json:"behaviors"     doc:"Each behavior that the API serves, and each internal behavior, in the order of the registry."`
	Areas        []CatalogArea      `                json:"areas"         doc:"Each area, in order, with its label."`
	Permissions  []types.Permission `                json:"permissions"   doc:"Each permission that osapi knows."                                                            render:"-"`
	Roles        []CatalogRole      `                json:"roles"         doc:"The built-in roles, in order, with the permissions of each."`
	Operations   []CatalogOperation `                json:"operations"    doc:"The permission of each operation that needs one, in the order of the operationIds."`
	StatusTones  map[string]string  `                json:"status_tones"  doc:"The tone of each status word, as the CLI colors it: ok, warn or err."                         render:"-"`
	CountTones   map[string]string  `                json:"count_tones"   doc:"The tone of a count above zero, by what it counts: ok, warn or err."                          render:"-"`
	HostFamilies []string           `                json:"host_families" doc:"The host platforms that gohai names. A method's runs_on has * for any other Linux."`
}

// CatalogArea is one area: a section of the Behaviors page.
type CatalogArea struct { //nolint:revive // the schema name is part of the API
	Name  string `json:"name"  doc:"The name of the area."  example:"networking"`
	Label string `json:"label" doc:"The label of the area." example:"Networking"`
}

// CatalogRole is one built-in role.
type CatalogRole struct { //nolint:revive // the schema name is part of the API
	Name        types.Role         `json:"name"        doc:"The name of the role, as a token names it." example:"operator"`
	Description string             `json:"description" doc:"Who the role is for."`
	Permissions []types.Permission `json:"permissions" doc:"The permissions of the role."                                  table:"-"`
}

// CatalogOperation is the permission of one operation.
type CatalogOperation struct { //nolint:revive // the schema name is part of the API
	ID         OperationID      `json:"id"         doc:"The operationId."`
	Permission types.Permission `json:"permission" doc:"The permission that the operation needs."`
}

// CatalogBehavior is one behavior, as the UI shows it. An internal behavior
// is in the catalog so a chain can name it, but it has no route: only
// another behavior calls it.
type CatalogBehavior struct { //nolint:revive // the schema name is part of the API
	Name      string          `json:"name"             doc:"The name of the behavior."                                         example:"dns"`
	Title     string          `json:"title"            doc:"The title of the behavior."                                        example:"DNS"`
	Area      string          `json:"area"             doc:"The area of the behavior, or empty for an internal behavior."      example:"networking"`
	Doc       string          `json:"doc"              doc:"What the behavior manages."                                                             table:"-"`
	Internal  bool            `json:"internal"         doc:"Only other behaviors call it: it has no route."`
	Parent    string          `json:"parent,omitempty" doc:"The behavior that the behavior is under, such as user for sshkey."`
	Platforms []string        `json:"platforms"        doc:"The platforms of the providers of the behavior."`
	Uses      []string        `json:"uses"             doc:"The behaviors that the behavior calls."`
	UsedBy    []string        `json:"used_by"          doc:"The behaviors that call the behavior."`
	Chain     string          `json:"chain"            doc:"The chain of the behavior, as osapi dsl chain list prints it."`
	Methods   []CatalogMethod `json:"methods"          doc:"The methods of the behavior."`
}

// CatalogMethod is one method of a behavior, as the UI shows and runs it.
type CatalogMethod struct { //nolint:revive // the schema name is part of the API
	Name       string           `json:"name"                  doc:"The name of the method."                                                                                                          example:"update"`
	Kind       types.MethodKind `json:"kind"                  doc:"What the method does to a host, by the verb of its permission: read, write or action."                                            example:"write"`
	Doc        string           `json:"doc"                   doc:"What the method does."`
	Long       string           `json:"long,omitempty"        doc:"The rest of what the method does."`
	HTTPMethod string           `json:"http_method,omitempty" doc:"The HTTP method of the route. An internal behavior has no route."                                                                 example:"PUT"`
	Path       string           `json:"path,omitempty"        doc:"The path of the route."`
	Permission types.Permission `json:"permission,omitempty"  doc:"The permission of the method. An internal behavior has none."`
	Platforms  []string         `json:"platforms"             doc:"The platforms that the method runs on: those that its On limits it to, else each platform of the behavior."`
	RunsOn     []string         `json:"runs_on"               doc:"The host platforms that the method runs on, as the agent decides, and * for a Linux host whose platform gohai has no family for."`
	Input      string           `json:"input"                 doc:"The input type and its fields, a required one marked *."`
	Output     string           `json:"output"                doc:"The output type and its fields."`
	Fields     []CatalogField   `json:"fields"                doc:"The fields of the input."`
	Columns    []CatalogColumn  `json:"columns"               doc:"The fields of the rows that the method answers, as the CLI's output reads them."`
	Calls      []CatalogCall    `json:"calls"                 doc:"The calls that the examples of the method make to other behaviors."`
	Examples   []CatalogExample `json:"examples"              doc:"The examples of the method."`
	Disruptive bool             `json:"disruptive,omitempty"  doc:"The method can interrupt a host, such as a reboot."`
	Disruption string           `json:"disruption,omitempty"  doc:"What the method interrupts."`
}

// CatalogField is one input field of a method: a field of the form of a
// step.
type CatalogField struct { //nolint:revive // the schema name is part of the API
	Name     string   `json:"name"              doc:"The json name of the field."                                                         example:"servers"`
	In       string   `json:"in"                doc:"Where the field goes in the request: path, query or body."                           example:"body"    validate:"oneof=path query body"`
	Kind     string   `json:"kind"              doc:"How a form takes the value: string, number, bool, list or object."                   example:"list"    validate:"oneof=string number bool list object"`
	Required bool     `json:"required"          doc:"The request needs the field."`
	Doc      string   `json:"doc"               doc:"What the field is."`
	Options  []string `json:"options,omitempty" doc:"The values that the field takes: those of its type with enum, or of its oneof rule."`
	Secret   bool     `json:"secret,omitempty"  doc:"The field is a secret."`
}

// CatalogColumn is one field of the rows that a method answers, as the
// CLI's output shows it: its label, and its render tag, such as bytes. A
// field that holds an object or a list of objects has the columns of the
// object, so the UI shows it as the CLI does, with no card of its own.
type CatalogColumn struct { //nolint:revive // the schema name is part of the API
	Name    string          `json:"name"              doc:"The json name of the field."                          example:"total"`
	Label   string          `json:"label"             doc:"The label of the field."                              example:"Total"`
	Doc     string          `json:"doc"               doc:"What the field is."`
	Render  string          `json:"render,omitempty"  doc:"The render tag of the field, such as bytes."          example:"bytes"`
	Kind    string          `json:"kind"              doc:"What the field holds: value, values, object or list." example:"value" validate:"oneof=value values object list"`
	Columns []CatalogColumn `json:"columns,omitempty" doc:"The columns of an object, or of a list of objects."`
}

// CatalogCall is one call that a method makes to another behavior, with the
// calls that it makes in turn, as the examples of the method record them.
type CatalogCall struct { //nolint:revive // the schema name is part of the API
	Behavior string        `json:"behavior"        doc:"The behavior that is called."   example:"package"`
	Method   string        `json:"method"          doc:"The method that is called."     example:"get"`
	Calls    []CatalogCall `json:"calls,omitempty" doc:"The calls that the call makes."`
}

// CatalogExample is one example of a method: what it does, its input and
// its output. A secret of an input is [redacted].
type CatalogExample struct { //nolint:revive // the schema name is part of the API
	Doc    string `json:"doc"              doc:"What the example does."`
	Input  any    `json:"input"            doc:"The input of the example."`
	Output any    `json:"output,omitempty" doc:"The output of the example."`
}
