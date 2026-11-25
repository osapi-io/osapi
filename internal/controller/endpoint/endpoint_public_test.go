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

package endpoint_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/apicontract"
	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/secret"
)

// letter is a name of a test, a closed set for the enum rule.
type letter string

func (letter) Values() []letter { return []letter{"a", "b"} }

func (v letter) Valid() bool { return v == "a" || v == "b" }

type EndpointPublicTestSuite struct {
	suite.Suite
}

var errMissing = errors.New("missing")

type noInput struct{}

type hostKey struct {
	Hostname string `json:"hostname" in:"path" param:"Hostname" validate:"required,min=1"`
	Key      string `json:"key"      in:"path" param:"NodeKey"  validate:"required,min=1"`
}

type hostOnly struct {
	Hostname string `json:"hostname" in:"path" param:"Hostname" validate:"required,min=1"`
}

type limited struct {
	Name string `json:"name" in:"query" validate:"omitempty,max=3"`
}

type unknownRule struct {
	Skip string `json:"skip" in:"query" validate:"nosuchrule"`
}

type boundsInput struct {
	Paths []string `json:"paths" validate:"required,min=1,max=3"`
	Dir   string   `json:"dir"   validate:"omitempty,startswith=/etc"`
}

type badBound struct {
	Count int `json:"count" validate:"min=x"`
}

type keyInput struct {
	Hostname string `json:"hostname"        in:"path"  validate:"required,min=1,max=255" doc:"The host."`
	Force    bool   `json:"force"           in:"query" validate:"required"`
	Limit    int    `json:"limit,omitempty" in:"query" validate:"omitempty,min=1"`
	Hidden   string `json:"-"`
	private  string //nolint:unused
}

type bodyInput struct {
	Name  letter   `json:"name"  validate:"required,enum"`
	Count uint     `json:"count"`
	Ratio float32  `json:"ratio"`
	Tags  []string `json:"tags"`
}

type optionalBody struct {
	Note string `json:"note"`
}

type Request struct {
	Reason string `json:"reason"`
}

type formInput struct {
	Name string `json:"name" in:"form" validate:"required"`
	File []byte `json:"file" in:"form" validate:"required"`
}

type bytesInput struct {
	Data []byte `json:"data"`
}

type namedBodyInput struct {
	Body Request `in:"body" validate:"required"`
}

type optionalNamedBodyInput struct {
	Body Request `in:"body"`
}

type Node struct {
	Name     string            `json:"name"`
	Seen     time.Time         `json:"seen,omitzero"`
	Labels   map[string]string `json:"labels,omitempty"`
	Facts    map[string]any    `json:"facts,omitempty"`
	Any      any               `json:"any,omitempty"`
	Parent   *Node             `json:"parent,omitempty"`
	Children []Node            `json:"children,omitempty"`
	Pair     [2]bool           `json:"pair,omitempty"`
	Inline   struct {
		Value int `json:"value"`
	} `json:"inline"`
	NoTag string
}

type badOut struct {
	C chan int `json:"c"`
}

type BadNamed struct {
	C chan int `json:"c"`
}

type deps struct{}

func handleNode(
	_ context.Context,
	_ deps,
	_ noInput,
) (Node, error) {
	return Node{}, nil
}

type underIn struct {
	Vault string `json:"vault" in:"path" validate:"required"`
	Key   string `json:"key"   in:"path" validate:"required"`
}

type vaultIn struct {
	Vault string `json:"vault" in:"path" validate:"required"`
}

type secretKeyIn struct {
	Vault string       `json:"vault" in:"path" validate:"required"`
	Token secret.Value `json:"token" in:"path" validate:"required"`
}

type twoKeys struct {
	A string `json:"a" in:"path"`
	B string `json:"b" in:"path"`
}

type badMap struct {
	M map[string]chan int `json:"m"`
}

type badSlice struct {
	S []chan int `json:"s"`
}

type badQuery struct {
	C chan int `json:"c" in:"query"`
}

type badBody struct {
	C chan int `json:"c"`
}

type badNamedBody struct {
	Body BadNamed `in:"body"`
}

// contract is the contract lines of g's spec.
func (s *EndpointPublicTestSuite) contract(
	g *endpoint.Group,
) []string {
	doc, err := endpoint.OpenAPI(g)
	s.Require().NoError(err)
	raw, err := json.Marshal(doc)
	s.Require().NoError(err)
	var m map[string]any
	s.Require().NoError(json.Unmarshal(raw, &m))

	return apicontract.Lines(m)
}

func (s *EndpointPublicTestSuite) TestUsesAndHandle() {
	m := endpoint.Get[noInput, Node]("list").Handle(handleNode)
	g := endpoint.Define("node", "The nodes", endpoint.Uses[deps](), m)

	s.Equal("endpoint_test.deps", g.Deps.String())
	s.NotNil(m.Handler())
	s.Nil(endpoint.Get[noInput, Node]("get").Handler())
	s.Same(m, g.Method("list"))
	s.Nil(g.Method("none"))

	e := endpoint.NotFound(errMissing, "gone")
	created := endpoint.Post[noInput, Node]("create").Status(http.StatusCreated).Errors(e)
	s.Equal(http.StatusCreated, created.SuccessStatus())
	s.Equal(http.StatusOK, m.SuccessStatus())
	s.Equal([]endpoint.Error{e}, created.Mapped())
}

func (s *EndpointPublicTestSuite) TestDeclared() {
	g := endpoint.Define("node", "Nodes",
		endpoint.Get[noInput, Node]("list"),
		endpoint.Post[keyInput, Node]("drain").AdminOnly(),
		endpoint.Get[noInput, Node]("trace").Permission("trace:read"),
		endpoint.Get[noInput, struct{}]("ping").Public(),
	)

	s.Equal([]endpoint.Declared{
		{Permission: "node:read", Where: "node list"},
		{Permission: "node:write", AdminOnly: true, Where: "node drain"},
		{Permission: "trace:read", Named: true, Where: "node trace"},
	}, g.Declared(), "a public method needs no permission")
}

func (s *EndpointPublicTestSuite) TestCommandHelp() {
	plain := endpoint.Get[noInput, Node]("get").Describe("Get a node").Explain("API words.")
	s.Equal("Get a node", plain.CommandShort())
	s.Equal("API words.", plain.CommandLong())
	s.Empty(plain.Examples())

	cli := endpoint.Get[noInput, Node]("list").
		Describe("List nodes").
		Command("List the nodes", "Shell words.").
		Example("").
		Example("--all")
	s.Equal("List the nodes", cli.CommandShort())
	s.Equal("Shell words.", cli.CommandLong())
	s.Equal([]endpoint.CommandExample{{Flags: ""}, {Flags: "--all"}}, cli.Examples())
	s.Empty(cli.PageNotes())
	s.Empty(cli.PageResult())

	paged := endpoint.Get[noInput, Node]("get").Notes("More in --json.").Result("Fails twice.")
	s.Equal("More in --json.", paged.PageNotes())
	s.Equal("Fails twice.", paged.PageResult())
}

func (s *EndpointPublicTestSuite) TestInternal() {
	m := endpoint.Get[noInput, Node]("get").Sub("trace")
	s.Empty(m.Saying())
	s.Equal("draining", endpoint.Post[noInput, struct{}]("drain").Says("draining").Saying())
	s.Equal(reflect.TypeFor[Node](), m.Output())
	s.Equal("trace", m.SubPath())
	s.Equal("vault",
		endpoint.Get[noInput, Node]("vaults").CommandSub("vault").Use("list").SubPath(),
		"the command's word, which the path does not have")
	s.Equal("get", m.CommandName())
	s.Equal("list", endpoint.Get[noInput, Node]("pending").Use("list").CommandName())

	probe := endpoint.Get[noInput, Node](
		"ready",
	).Answers(http.StatusServiceUnavailable, "Not ready.")
	s.Equal([]endpoint.Answered{{Code: http.StatusServiceUnavailable, Doc: "Not ready."}},
		probe.Answering())
	s.EqualError(endpoint.Answer(http.StatusServiceUnavailable), "Service Unavailable")

	cause := errors.New("dial tcp 10.0.0.9:16686: refused")
	err := endpoint.Internal("failed to read the trace", cause)
	s.EqualError(err, "failed to read the trace")
	s.ErrorIs(err, cause)
}

func (s *EndpointPublicTestSuite) TestErrorf() {
	err := endpoint.Errorf(errMissing, "no node named %q", "web-01")

	s.EqualError(err, `no node named "web-01"`)
	s.ErrorIs(err, errMissing)
}

func (s *EndpointPublicTestSuite) TestDefine() {
	g := endpoint.Define("node", "The nodes")
	s.Equal("node", g.Title)
	s.Equal("node", g.Tag)

	g = endpoint.Define("node", "The nodes",
		endpoint.Title("Node API"),
		endpoint.Tag("node_operations", "Node", "Operations on nodes."),
		endpoint.Get[noInput, Node]("list"),
	)
	s.Equal("Node API", g.Title)
	s.Equal([]string{"node_operations", "Node", "Operations on nodes."},
		[]string{g.Tag, g.TagDisplay, g.TagDoc})
	s.Len(g.Methods, 1)
}

func (s *EndpointPublicTestSuite) TestPathPermissionAndID() {
	tests := []struct {
		name   string
		method *endpoint.Method
		path   string
		perm   string
		id     string
	}{
		{
			name:   "when a GET has no key",
			method: endpoint.Get[noInput, Node]("list"),
			path:   "/api/node",
			perm:   "node:read",
			id:     "GetNodeList",
		},
		{
			name:   "when a GET has a sub-path and a key",
			method: endpoint.Get[keyInput, Node]("get").Sub("trace"),
			path:   "/api/node/trace/{hostname}",
			perm:   "node:read",
			id:     "GetNodeTraceGet",
		},
		{
			name: "when a POST has an action, an ID and a permission",
			method: endpoint.Post[keyInput, Node](
				"drain",
			).Action("drain").
				ID("drainNode").
				Permission("agent:write"),
			path: "/api/node/{hostname}/drain",
			perm: "agent:write",
			id:   "drainNode",
		},
		{
			name:   "when a GET has a key of two fields, in the order Key names",
			method: endpoint.Get[twoKeys, Node]("get").Key("b", "a"),
			path:   "/api/node/{b}/{a}",
			perm:   "node:read",
			id:     "GetNodeGet",
		},
		{
			name:   "when a list names only the field that holds the keys",
			method: endpoint.Get[keyInput, Node]("list").Key("hostname"),
			path:   "/api/node/{hostname}",
			perm:   "node:read",
			id:     "GetNodeList",
		},
		{
			name:   "when a PUT has the default permission",
			method: endpoint.Put[keyInput, Node]("update"),
			path:   "/api/node/{hostname}",
			perm:   "node:write",
			id:     "PutNodeUpdate",
		},
		{
			name:   "when a DELETE is public",
			method: endpoint.Delete[keyInput, Node]("remove").Public(),
			path:   "/api/node/{hostname}",
			perm:   "",
			id:     "DeleteNodeRemove",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			g := endpoint.Define("node", "The nodes", tc.method)

			s.Equal(tc.path, g.Path(tc.method))
			s.Equal(tc.perm, string(g.Permission(tc.method)))
			s.Equal(tc.id, tc.method.OperationID(g))
		})
	}
}

func (s *EndpointPublicTestSuite) TestRegistry() {
	g := endpoint.Register(endpoint.Define("zzregistry", "A test group"))

	s.Same(g, endpoint.Lookup("zzregistry"))
	s.Nil(endpoint.Lookup("zznone"))
	s.Contains(endpoint.All(), g)
	s.PanicsWithValue(`endpoint: group "zzregistry" registered twice`, func() {
		endpoint.Register(endpoint.Define("zzregistry", "Again"))
	})
	endpoint.Register(endpoint.Define("zzaregistry", "Sorts first"))
	all := endpoint.All()
	for i := 1; i < len(all); i++ {
		s.Less(all[i-1].Name, all[i].Name)
	}
}

func (s *EndpointPublicTestSuite) TestOpenAPI() {
	tests := []struct {
		name    string
		methods []endpoint.Option
		want    []string
	}{
		{
			name: "when a method has a key, a query, errors and a status",
			methods: []endpoint.Option{
				endpoint.Post[keyInput, Node]("drain").
					Action("drain").
					Describe("Drain a node").
					Explain("A drained node takes no new jobs.").
					Returns("Drained.").
					Status(http.StatusCreated).
					Errors(
						endpoint.NotFound(errMissing, "No such node."),
						endpoint.Conflict(errMissing, "Already drained."),
						endpoint.Unavailable(errMissing, "Not ready."),
						endpoint.Invalid(errMissing, "Nothing to drain."),
						endpoint.Forbidden(errMissing, "Not this one."),
					),
			},
			want: []string{
				"op POST /api/node/{hostname}/drain id=PostNodeDrain tags=[node] " +
					"auth=BearerAuth:node:write " +
					"params=[path:hostname!:string query:force!:boolean query:limit:integer] body= " +
					"responses=[201=application/json:Node 400=application/json:ErrorResponse " +
					"401=application/json:ErrorResponse 403=application/json:ErrorResponse " +
					"404=application/json:ErrorResponse 409=application/json:ErrorResponse " +
					"500=application/json:ErrorResponse 503=application/json:ErrorResponse]",
			},
		},
		{
			name: "when the input is not a struct, it has no parameters",
			methods: []endpoint.Option{
				endpoint.Get[string, Node]("raw"),
			},
			want: []string{
				"op GET /api/node id=GetNodeRaw tags=[node] auth=BearerAuth:node:read params=[] body= " +
					"responses=[200=application/json:Node 401=application/json:ErrorResponse " +
					"403=application/json:ErrorResponse 500=application/json:ErrorResponse]",
			},
		},
		{
			name: "when a public method has no input and answers 204",
			methods: []endpoint.Option{
				endpoint.Get[noInput, struct{}]("ping").Public().Status(http.StatusNoContent),
				endpoint.Get[noInput, Node]("ready").Sub("ready").Public().
					Answers(http.StatusServiceUnavailable, "Not ready."),
			},
			want: []string{
				"op GET /api/node/ready id=GetNodeReadyReady tags=[node] auth=public params=[] body= " +
					"responses=[200=application/json:Node 500=application/json:ErrorResponse " +
					"503=application/json:Node]",
				"op GET /api/node id=GetNodePing tags=[node] auth=public params=[] body= " +
					"responses=[204=- 500=application/json:ErrorResponse]",
			},
		},
		{
			name: "when the body is the input's other fields",
			methods: []endpoint.Option{
				endpoint.Post[bodyInput, struct {
					OK bool `json:"ok"`
				}]("create"),
				endpoint.Put[optionalBody, struct{}]("note").Sub("note"),
			},
			want: []string{
				"op POST /api/node id=PostNodeCreate tags=[node] auth=BearerAuth:node:write params=[] " +
					"body=application/json:{count:integer,name!:letter,ratio:number,tags:[string]}! " +
					"responses=[200=application/json:{ok!:boolean} 400=application/json:ErrorResponse " +
					"401=application/json:ErrorResponse 403=application/json:ErrorResponse " +
					"500=application/json:ErrorResponse]",
				"op PUT /api/node/note id=PutNodeNoteNote tags=[node] auth=BearerAuth:node:write params=[] " +
					"body=application/json:{note:string} " +
					"responses=[200=application/json:{} 400=application/json:ErrorResponse " +
					"401=application/json:ErrorResponse 403=application/json:ErrorResponse " +
					"500=application/json:ErrorResponse]",
			},
		},
		{
			name: "when the body is a form, a []byte is the file",
			methods: []endpoint.Option{
				endpoint.Post[formInput, Node]("upload").Sub("upload"),
				endpoint.Post[bytesInput, Node]("raw").Sub("raw"),
			},
			want: []string{
				"op POST /api/node/upload id=PostNodeUploadUpload tags=[node] auth=BearerAuth:node:write " +
					"params=[] body=multipart/form-data:{file!:string/binary,name!:string}! " +
					"responses=[200=application/json:Node 400=application/json:ErrorResponse " +
					"401=application/json:ErrorResponse 403=application/json:ErrorResponse " +
					"500=application/json:ErrorResponse]",
				"op POST /api/node/raw id=PostNodeRawRaw tags=[node] auth=BearerAuth:node:write " +
					"params=[] body=application/json:{data:string/byte} " +
					"responses=[200=application/json:Node 400=application/json:ErrorResponse " +
					"401=application/json:ErrorResponse 403=application/json:ErrorResponse " +
					"500=application/json:ErrorResponse]",
			},
		},
		{
			name: "when the body is a named type",
			methods: []endpoint.Option{
				endpoint.Post[namedBodyInput, *Node]("retry"),
				endpoint.Post[optionalNamedBodyInput, *Node]("redo").Sub("redo"),
			},
			want: []string{
				"op POST /api/node id=PostNodeRetry tags=[node] auth=BearerAuth:node:write params=[] " +
					"body=application/json:Request! " +
					"responses=[200=application/json:Node 400=application/json:ErrorResponse " +
					"401=application/json:ErrorResponse 403=application/json:ErrorResponse " +
					"500=application/json:ErrorResponse]",
				"schema Request.reason! = string",
				"op POST /api/node/redo id=PostNodeRedoRedo tags=[node] auth=BearerAuth:node:write " +
					"params=[] body=application/json:Request " +
					"responses=[200=application/json:Node 400=application/json:ErrorResponse " +
					"401=application/json:ErrorResponse 403=application/json:ErrorResponse " +
					"500=application/json:ErrorResponse]",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			opts := append([]endpoint.Option{endpoint.Tag("node", "Node", "Nodes.")}, tc.methods...)
			lines := s.contract(endpoint.Define("node", "The nodes", opts...))

			for _, w := range tc.want {
				s.Contains(lines, w)
			}
		})
	}
}

type Sample struct {
	Count int     `json:"count" example:"4"`
	Load  float64 `json:"load"  example:"0.32"`
	On    bool    `json:"on"    example:"true"`
	Name  string  `json:"name"  example:"web-01"`
	Bad   int     `json:"bad"   example:"many"`
	Badf  float64 `json:"badf"  example:"x"`
	Badb  bool    `json:"badb"  example:"maybe"`
	ID    string  `json:"id"                     pattern:"^[0-9a-f]{32}$"`
	Key   string  `json:"key"                                             validate:"uuid"`
	Took  int64   `json:"took"`
	Limit int     `json:"limit"                                           validate:"min=1,max=100" default:"20"`
	Ratio float64 `json:"ratio"                                           validate:"min=0"`
	Small float32 `json:"small"`
	// An object's and a list's examples are their JSON.
	Counts map[string]int `json:"counts" example:"{\"ok\":5}"`
	Tags   []string       `json:"tags"   example:"[\"web\"]"`
	Odd    map[string]int `json:"odd"    example:"not json"`
}

func (s *EndpointPublicTestSuite) TestLocation() {
	m := endpoint.Post[noInput, Node]("make").Status(http.StatusCreated).Location("name")
	s.Equal("name", m.LocationField())

	doc, err := endpoint.OpenAPI(endpoint.Define("node", "The nodes", m))
	s.Require().NoError(err)
	created := doc.Paths.Find("/api/node").Post.Responses.Value("201").Value
	s.Contains(created.Headers, "Location")
}

func (s *EndpointPublicTestSuite) TestExamplesAndDocs() {
	g := endpoint.Define("node", "The nodes",
		endpoint.Doc[Sample]("A sample."),
		endpoint.Example[Sample]("name", `{"host":"web-01"}`),
		endpoint.Example[Sample]("nosuchfield", "x"),
		endpoint.Get[noInput, Sample]("get"),
	)
	doc, err := endpoint.OpenAPI(g)
	s.Require().NoError(err)

	sample := doc.Components.Schemas["Sample"].Value
	s.Equal("A sample.", sample.Description)
	s.Equal(int64(4), sample.Properties["count"].Value.Example)
	s.InEpsilon(0.32, sample.Properties["load"].Value.Example, 0.0001)
	s.Equal(true, sample.Properties["on"].Value.Example)
	s.Equal(`{"host":"web-01"}`, sample.Properties["name"].Value.Example)
	s.Equal("many", sample.Properties["bad"].Value.Example)
	s.Equal(map[string]any{"ok": float64(5)}, sample.Properties["counts"].Value.Example,
		"an object's example is an object, not a string")
	s.Equal([]any{"web"}, sample.Properties["tags"].Value.Example)
	s.Equal("not json", sample.Properties["odd"].Value.Example)
	s.Equal("x", sample.Properties["badf"].Value.Example)
	s.Equal("maybe", sample.Properties["badb"].Value.Example)
	s.Equal("^[0-9a-f]{32}$", sample.Properties["id"].Value.Pattern)
	s.Equal("double", sample.Properties["load"].Value.Format)
	s.Empty(sample.Properties["small"].Value.Format)
	s.Equal("uuid", sample.Properties["key"].Value.Format)
	s.Equal("int64", sample.Properties["took"].Value.Format)
	limit := sample.Properties["limit"].Value
	s.Empty(limit.Format)
	s.InDelta(1.0, *limit.Min, 0)
	s.InDelta(100.0, *limit.Max, 0)
	s.Equal(int64(20), limit.Default)
	s.InDelta(0.0, *sample.Properties["ratio"].Value.Min, 0)
}

// shade is a closed set: a named string type with Values and Valid.
type shade string

func (shade) Values() []shade { return []shade{"red", "blue"} }

func (c shade) Valid() bool { return c == "red" || c == "blue" }

// palette holds shades: in a field, in a list and in a map.
type palette struct {
	Main   shade            `json:"main"    doc:"The main shade."`
	Others []shade          `json:"others"  doc:"The other shades."`
	ByName map[string]shade `json:"by_name" doc:"The shades by name."`
}

// paint is an input with a shade, which the enum rule checks.
type paint struct {
	Shade shade `json:"shade" validate:"required,enum"`
}

// badPaint lists the values of a shade again, with oneof.
type badPaint struct {
	Shade shade `json:"shade" validate:"required,oneof=red blue"`
}

// TestEnum makes a closed set a schema of the components, once, with the
// values of its Values; the enum rule is its rule, and oneof is refused.
func (s *EndpointPublicTestSuite) TestEnum() {
	g := endpoint.Define("paint", "The paints",
		endpoint.Doc[shade]("A shade."),
		endpoint.Get[noInput, palette]("get"),
		endpoint.Post[paint, palette]("create"),
	)
	doc, err := endpoint.OpenAPI(g)
	s.Require().NoError(err)

	sh := doc.Components.Schemas["shade"].Value
	s.Equal("A shade.", sh.Description)
	s.Equal([]any{"red", "blue"}, sh.Enum)
	s.True(sh.Type.Is("string"))
	p := doc.Components.Schemas["palette"].Value
	s.Equal("#/components/schemas/shade", p.Properties["main"].Ref)
	s.Equal("#/components/schemas/shade", p.Properties["others"].Value.Items.Ref)
	s.Equal("#/components/schemas/shade",
		p.Properties["by_name"].Value.AdditionalProperties.Schema.Ref)

	_, err = endpoint.OpenAPI(endpoint.Define("paint", "The paints",
		endpoint.Post[badPaint, palette]("create")))
	s.ErrorContains(err, "the type has Values; use validate enum")
}

// tint is a closed set that describes itself, as a type of pkg/sdk/types
// does.
type tint string

func (tint) Values() []tint { return []tint{"warm", "cool"} }

func (t tint) Valid() bool { return t == "warm" || t == "cool" }

func (tint) Doc() string { return "A tint." }

// tinted holds a tint.
type tinted struct {
	Tint tint `json:"tint" validate:"required,enum"`
}

// TestEnumDoc takes a closed set's description from its Doc when the group
// gives none, so two groups that use it describe it the same way.
func (s *EndpointPublicTestSuite) TestEnumDoc() {
	doc, err := endpoint.OpenAPI(endpoint.Define("tint", "The tints",
		endpoint.Post[tinted, tinted]("create")))
	s.Require().NoError(err)

	s.Equal("A tint.", doc.Components.Schemas["tint"].Value.Description)
}

// TestBodyValidateTags holds a body field to its whole validate tag: the
// items of an array, a pattern from startswith, and the tag in the extension.
// TestPrefix puts a group under a prefix whose {field} is a shared path
// parameter of each method, and keeps the method's own key after the name.
func (s *EndpointPublicTestSuite) TestPrefix() {
	g := endpoint.Define("sysctl", "Kernel parameters", endpoint.Prefix("/fleet/node/{hostname}/"),
		endpoint.Area(area.Config),
		endpoint.Get[hostOnly, Node]("list"),
		endpoint.Get[hostKey, Node]("get"),
	)

	s.Equal("fleet/node/{hostname}", g.Prefix)
	s.Equal(area.Config, g.Area)
	s.Equal("/api/fleet/node/{hostname}/sysctl", g.Path(g.Method("list")))
	s.Equal("/api/fleet/node/{hostname}/sysctl/{key}", g.Path(g.Method("get")))
	s.Empty(g.Method("list").KeyField(), "the prefix's field is not the key")
	s.Empty(g.Method("list").Keys())
	s.Equal("key", g.Method("get").KeyField())
	s.Equal([]string{"key"}, g.Method("get").Keys())
	two := endpoint.Get[underIn, Node]("get").Key("vault", "key")
	s.Equal([]string{"vault", "key"}, two.Keys())
	s.Equal("key", two.KeyField(), "the last field names what the method acts on")
	s.Empty(endpoint.Validate(g))

	doc, err := endpoint.OpenAPI(g)
	s.Require().NoError(err)
	s.Equal([]string{"Hostname", "NodeKey"}, func() []string {
		names := make([]string, 0, len(doc.Components.Parameters))
		for name := range doc.Components.Parameters {
			names = append(names, name)
		}
		slices.Sort(names)

		return names
	}())
	get := doc.Paths.Value("/api/fleet/node/{hostname}/sysctl/{key}").Get
	s.Equal("#/components/parameters/Hostname", get.Parameters[0].Ref)
	s.Equal("#/components/parameters/NodeKey", get.Parameters[1].Ref)
	s.Equal("config", get.Extensions[endpoint.AreaExtension], "the operation has its group's area")
	s.Equal("sysctl", doc.Tags[0].Extensions[endpoint.GroupExtension], "the tag names its group")

	missing := endpoint.Define(
		"sysctl",
		"Kernel parameters",
		endpoint.Prefix("next/node/{hostname}"),
		endpoint.Get[noInput, Node]("list"),
	)
	problems := endpoint.Validate(missing)
	msgs := make([]string, 0, len(problems))
	for _, p := range problems {
		msgs = append(msgs, p.Message)
	}
	s.Contains(msgs, "the prefix names {hostname}, and the input has no path field hostname.")
}

// TestCommandPath is the words of a group's command: the fixed segments of
// its prefix, then its name; and CommandNotes is its page's Notes.
func (s *EndpointPublicTestSuite) TestCommandPath() {
	g := endpoint.Define("sysctl", "Kernel parameters",
		endpoint.Prefix("next/node/{hostname}"),
		endpoint.CommandNotes("See the sysctl page."),
	)
	s.Equal([]string{"next", "node", "sysctl"}, g.CommandPath())
	s.Equal("See the sysctl page.", g.CommandNotes)
	s.Equal([]string{"job"}, endpoint.Define("job", "Jobs").CommandPath())

	m := endpoint.Get[noInput, Node]("list").Example("--all", "Every one", "of them").Example("")
	s.Equal([]endpoint.CommandExample{{Flags: "--all", Doc: "Every one of them"}, {}}, m.Examples())
}

// TestCommandLayer is the group and its methods as the command layer reads
// them.
func (s *EndpointPublicTestSuite) TestCommandLayer() {
	g := endpoint.Define("sysctl", "Kernel parameters",
		endpoint.Area(area.Config),
		endpoint.CommandNotes("See the sysctl page."),
		endpoint.CommandLong("Each parameter is a file under sysctl.d."),
		endpoint.SubDoc("files", "Read the files"),
		endpoint.SubDoc("keys", "Read the keys"),
		endpoint.Get[noInput, Node]("list"),
		endpoint.Put[noInput, Node]("set").Permission("sysctl:admin"),
	)
	s.Equal("sysctl", g.GroupName())
	s.Equal("Kernel parameters", g.GroupDoc())
	s.Equal(area.Config, g.GroupArea())
	s.Equal("See the sysctl page.", g.GroupNotes())
	s.Equal("Each parameter is a file under sysctl.d.", g.GroupLong())
	s.Equal("Read the files", g.SubDoc("files"))
	s.Equal("Read the keys", g.SubDoc("keys"))
	s.Empty(g.SubDoc("other"))

	methods := g.Commands()
	s.Require().Len(methods, 2)
	s.Equal("list", methods[0].MethodName())
	s.Equal(reflect.TypeFor[noInput](), methods[0].Input())
	s.Equal(types.Permission("sysctl:read"), g.CommandPermission(methods[0]))
	s.Equal(types.Permission("sysctl:admin"), g.CommandPermission(methods[1]))
}

func (s *EndpointPublicTestSuite) TestRefusals() {
	m := endpoint.Get[limited, Node]("get").Fails("a name over 3 characters", limited{Name: "abcd"})

	s.Equal([]endpoint.Refusal{
		{Doc: "a name over 3 characters", Input: limited{Name: "abcd"}},
	}, m.Refusals())
}

func (s *EndpointPublicTestSuite) TestBodyValidateTags() {
	doc, err := endpoint.OpenAPI(endpoint.Define("node", "The nodes",
		endpoint.Post[boundsInput, Node]("make")))
	s.Require().NoError(err)

	body := doc.Paths.Value("/api/node").Post.RequestBody.Value.
		Content.Get("application/json").Schema.Value
	paths := body.Properties["paths"].Value
	s.Equal(uint64(1), paths.MinItems)
	s.Equal(uint64(3), *paths.MaxItems)
	s.Equal(map[string]any{
		"x-oapi-codegen-extra-tags": map[string]string{"validate": "required,min=1,max=3"},
	}, paths.Extensions)
	s.Equal("^/etc", body.Properties["dir"].Value.Pattern)
}

func (s *EndpointPublicTestSuite) TestParameterValidateTags() {
	doc, err := endpoint.OpenAPI(endpoint.Define("node", "The nodes",
		endpoint.Get[keyInput, Node]("get")))
	s.Require().NoError(err)

	params := doc.Paths.Value("/api/node/{hostname}").Get.Parameters
	s.Equal(map[string]any{
		"x-oapi-codegen-extra-tags": map[string]string{"validate": "required,min=1,max=255"},
	}, params[0].Value.Extensions)
	s.Equal(map[string]any{
		"x-oapi-codegen-extra-tags": map[string]string{"validate": "required"},
	}, params[1].Value.Extensions)
}

func (s *EndpointPublicTestSuite) TestOpenAPISchemas() {
	lines := s.contract(endpoint.Define("node", "The nodes", endpoint.Get[noInput, Node]("get")))

	s.Subset(lines, []string{
		"schema Node.NoTag! = string",
		"schema Node.any = ",
		"schema Node.children = [Node]",
		"schema Node.facts = map[any]",
		"schema Node.inline! = {value!:integer}",
		"schema Node.labels = map[string]",
		"schema Node.name! = string",
		"schema Node.pair = [boolean]",
		"schema Node.parent = Node",
		"schema Node.seen = string/date-time",
	})
}

func (s *EndpointPublicTestSuite) TestOpenAPIErrors() {
	tests := []struct {
		name    string
		method  *endpoint.Method
		wantErr string
	}{
		{
			name:    "when the response has a type JSON cannot encode",
			method:  endpoint.Get[noInput, badOut]("get"),
			wantErr: "kind chan has no schema",
		},
		{
			name:    "when a parameter's validate rule has no schema",
			method:  endpoint.Get[unknownRule, Node]("get"),
			wantErr: `validate tag "nosuchrule" has no schema`,
		},
		{
			name:    "when a body field's bound is not a number",
			method:  endpoint.Post[badBound, Node]("make"),
			wantErr: "validate min=x: not a number",
		},
		{
			name:    "when a named type has a field JSON cannot encode",
			method:  endpoint.Get[noInput, BadNamed]("get"),
			wantErr: "BadNamed: C",
		},
		{
			name:    "when a map's values cannot be encoded",
			method:  endpoint.Get[noInput, badMap]("get"),
			wantErr: "kind chan has no schema",
		},
		{
			name:    "when a slice's items cannot be encoded",
			method:  endpoint.Get[noInput, badSlice]("get"),
			wantErr: "kind chan has no schema",
		},
		{
			name:    "when a query parameter cannot be encoded",
			method:  endpoint.Get[badQuery, Node]("get"),
			wantErr: "C: type chan int",
		},
		{
			name:    "when a body field cannot be encoded",
			method:  endpoint.Post[badBody, Node]("create"),
			wantErr: "C: type chan int",
		},
		{
			name:    "when a named body cannot be encoded",
			method:  endpoint.Post[badNamedBody, Node]("create"),
			wantErr: "BadNamed",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := endpoint.OpenAPI(endpoint.Define("node", "The nodes", tc.method))

			s.ErrorContains(err, tc.wantErr)
		})
	}
}

// sentEmptyIn has an optional body field with no json omitempty, a path
// field, which is not in the body, and an optional field that is omitted.
type sentEmptyIn struct {
	ID   string `json:"id"             in:"path" validate:"required"`
	Note string `json:"note"                     validate:"omitempty"`
	Kept string `json:"kept,omitempty"           validate:"omitempty"`
	Skip string `json:"-"                        validate:"omitempty"`
}

// sentEmptyBodyIn names its body with in:"body", whose field is sent empty.
type sentEmptyBodyIn struct {
	ID   string        `json:"id" in:"path" validate:"required"`
	Body sentEmptyBody `          in:"body" validate:"required"`
}

type sentEmptyBody struct {
	Reason string `json:"reason" validate:"omitempty"`
}

// sentEmptyQuery is a read's query, which a client leaves out when empty.
type sentEmptyQuery struct {
	Note string `json:"note" in:"query" validate:"omitempty"`
}

func (s *EndpointPublicTestSuite) TestValidate() {
	tests := []struct {
		name  string
		group *endpoint.Group
		want  []string
	}{
		{
			name:  "when the group is valid",
			group: endpoint.Define("node", "The nodes", endpoint.Get[noInput, Node]("list")),
		},
		{
			name: "when a group that Uses a type has handlers of the right shape",
			group: endpoint.Define("node", "The nodes",
				endpoint.Uses[deps](),
				endpoint.Get[noInput, Node]("list").Handle(handleNode),
			),
		},
		{
			name: "when handlers are missing, wrong or without Uses",
			group: endpoint.Define("node", "The nodes",
				endpoint.Uses[deps](),
				endpoint.SubDoc("one", "One node"),
				endpoint.Get[noInput, Node]("list"),
				endpoint.Get[noInput, Node]("get").Sub("one").Handle(func() {}),
			),
			want: []string{
				"the group Uses endpoint_test.deps, and the method has no handler.",
				"the handler is func(); it must be func(context.Context, endpoint_test.deps, endpoint_test.noInput) (endpoint_test.Node, error).",
			},
		},
		{
			name: "when an input field's in tag is not path, query, form or body",
			group: endpoint.Define("node", "The nodes",
				endpoint.Get[badInIn, Node]("list"),
			),
			want: []string{
				`input field q is tagged in:"qeury"; the values are path, query, form and body.`,
			},
		},
		{
			name: "when Key names each path field once and in order, it passes",
			group: endpoint.Define("node", "The nodes",
				endpoint.Get[underIn, Node]("get").Key("vault", "key"),
				endpoint.Get[vaultIn, Node]("list").Key("vault"),
			),
		},
		{
			name: "when Key breaks the key rule, it says how",
			group: endpoint.Define("node", "The nodes",
				endpoint.Get[underIn, Node]("get"),
				endpoint.Delete[underIn, Node]("delete").Key("vault", "nope"),
				endpoint.Put[underIn, Node]("update").Key("vault", "vault"),
				endpoint.Post[underIn, Node]("create").Key("key"),
				endpoint.Post[secretKeyIn, Node]("seal").Sub("seal").Key("vault", "token"),
			),
			want: []string{
				"the input has more than one path field besides the prefix; name each in order with .Key.",
				"Key names {nope}, and the input has no path field nope.",
				"path field key is not in Key or the prefix.",
				"Key names {vault} twice.",
				"path field key is not in Key or the prefix.",
				"path field vault is not in Key or the prefix.",
				"input field token is a secret in the path; a URL goes in logs.",
				"Key names {token}, which is a secret; a URL goes in logs.",
			},
		},
		{
			name: "when a secret is in the URL or on a span",
			group: endpoint.Define("node", "The nodes",
				endpoint.Get[secretInURLIn, Node]("get"),
			),
			want: []string{
				"input field id is a secret in the path; a URL goes in logs.",
				"input field token is a secret in the query; a URL goes in logs.",
				`input field token is a secret and tagged telemetry:"attr"; a secret is never recorded on a span.`,
			},
		},
		{
			name: "when an example names a flag that the command does not have",
			group: endpoint.Define("node", "The nodes",
				endpoint.SubDoc("one", "One node"),
				endpoint.Get[exampleIn, Node]("list").
					Example("--hostname web-01 --limit=5 -T _all --json").
					Example("--hostnme web-01"),
				endpoint.Post[exampleBodyIn, Node]("create").Sub("one").
					Example("--id a --reason why").
					Example("--resaon why"),
			),
			want: []string{
				`the example "--hostnme web-01" names --hostnme, which the command does not have.`,
				`the example "--resaon why" names --resaon, which the command does not have.`,
			},
		},
		{
			name: "when an error has a reason that types.Reason does not list",
			group: endpoint.Define("node", "The nodes",
				endpoint.Get[noInput, Node]("list").Errors(
					endpoint.Conflict(errors.New("x"), "A conflict").Because("not_manged"),
				),
			),
			want: []string{
				`an error of 409 has the reason "not_manged", which is not one of the values of types.Reason.`,
			},
		},
		{
			name: "when Location or Example names a field that its type does not have",
			group: endpoint.Define("node", "The nodes",
				endpoint.Example[Node]("nmae", "web-01"),
				endpoint.Example[Node]("name", "web-01"),
				endpoint.Post[noInput, *Node]("create").Location("nmae"),
				endpoint.Post[noInput, Node]("update").Sub("one").Location("name"),
				endpoint.SubDoc("one", "One node"),
			),
			want: []string{
				"Example names nmae, which is not a json field of endpoint_test.Node.",
				"Location names nmae, which is not a json field of *endpoint_test.Node.",
			},
		},
		{
			name: "when Prints gives an example another type, or comes before any Example",
			group: endpoint.Define(
				"node",
				"The nodes",
				endpoint.Get[noInput, Node]("list").Prints(Node{}).Example("").Prints("web-01"),
				endpoint.Get[noInput, Node](
					"get",
				).Sub("one").
					Example("").
					Prints(Node{Name: "web-01"}).
					Example("--json"),
				endpoint.SubDoc("one", "One node"),
			),
			want: []string{
				"Prints comes before any Example; it gives the last Example its output.",
				`Prints gives the example "" a string, not a endpoint_test.Node.`,
			},
		},
		{
			name: "when a method has a handler and the group Uses no type",
			group: endpoint.Define("node", "The nodes",
				endpoint.Get[noInput, Node]("list").Handle(handleNode),
			),
			want: []string{"the method has a handler, and the group Uses no type."},
		},
		{
			name: "when an optional body field is sent empty, in the input or in its body type",
			group: endpoint.Define("node", "The nodes",
				endpoint.SubDoc("one", "One node"),
				endpoint.Post[sentEmptyIn, Node]("create"),
				endpoint.Post[sentEmptyBodyIn, Node]("reject").Sub("one"),
				endpoint.Get[sentEmptyQuery, Node]("list"),
			),
			want: []string{
				"optional body field note is sent as its zero value when it is empty.",
				"optional body field reason is sent as its zero value when it is empty.",
			},
		},
		{
			name:  "when the group has a bad name and no methods",
			group: endpoint.Define("Node", "The nodes"),
			want: []string{
				`group name "Node" is not lowercase letters and digits.`,
				"the group has no methods.",
			},
		},
		{
			name: "when the group's area is not one of area.All",
			group: endpoint.Define("node", "The nodes", endpoint.Area("kitchen"),
				endpoint.Get[noInput, Node]("list")),
			want: []string{
				`area "kitchen" is not one of services, software, config, node, logs, ` +
					`networking, security, containers, objectstore, secrets, command, hardware, audit, fleet.`,
			},
		},
		{
			name: "when a Sub segment makes a command with no SubDoc, it says so",
			group: endpoint.Define("node", "The nodes",
				endpoint.SubDoc("files", "Read the files"),
				endpoint.Get[noInput, Node]("get").Sub("trace"),
				endpoint.Get[noInput, Node]("list").Sub("files"),
				endpoint.Get[noInput, Node]("status").Sub("status"),
			),
			want: []string{`the Sub segment "trace" has no SubDoc.`},
		},
		{
			name: "when a CommandSub word has no SubDoc, or is not lowercase, it says so",
			group: endpoint.Define("node", "The nodes",
				endpoint.Get[noInput, Node]("vaults").CommandSub("vault").Use("list"),
				endpoint.Get[noInput, Node]("disks").Sub("disks").CommandSub("Disk").Use("list"),
			),
			want: []string{
				`the Sub segment "vault" has no SubDoc.`,
				`path segment "Disk" is not lowercase letters and digits.`,
				`the Sub segment "Disk" has no SubDoc.`,
			},
		},
		{
			name: "when methods break the rules",
			group: endpoint.Define("node", "The nodes",
				endpoint.Get[noInput, Node]("List"),
				endpoint.Get[noInput, Node]("list").ID("GetNodeList"),
				endpoint.Get[noInput, Node]("list").Sub("Bad"),
				endpoint.Post[noInput, Node]("act").Action("go"),
				endpoint.Get[twoKeys, Node]("two"),
				endpoint.Get[noInput, badOut]("bad").Sub("bad"),
				endpoint.Get[noInput, Node]("lower").Sub("lower").ID("getLower"),
				endpoint.Get[limited, Node]("short").Sub("short").Fails("not the input", "x"),
				endpoint.Get[limited, Node]("refused").Sub("refused").
					Fails("a name over 3 characters", limited{Name: "abcd"}),
			),
			want: []string{
				`method name "List" is not lowercase letters and digits.`,
				"operationId GetNodeList is also the ID of List.",
				"GET /api/node is also the route of List.",
				"two methods have the name list.",
				`path segment "Bad" is not lowercase letters and digits.`,
				`the Sub segment "Bad" has no SubDoc.`,
				"the method has an action but no key.",
				"the input has more than one path field besides the prefix; name each in order with .Key.",
				"input field a has no validate tag.",
				"input field b has no validate tag.",
				"operationId getLower starts with a small letter.",
				`the .Fails input "not the input" is string, not endpoint_test.limited.`,
				"no .Fails input breaks max=3 on Name.",
				"the spec does not build: node bad: response: badOut: C: type chan int: kind chan has no schema.",
			},
		},
	}

	seen := map[string]bool{}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Each group but the one about its area has one, so the other
			// rules are what the case reads.
			if tc.group.Area == "" {
				tc.group.Area = area.Fleet
			}
			var got []string
			for _, p := range endpoint.Validate(tc.group) {
				got = append(got, p.Message)
				seen[p.Rule] = true
				s.Equal(endpoint.Checks.Fix(p.Rule), p.Fix, "every rule has a check")
				s.NotEmpty(p.Fix)
				s.Equal(p.Message+" Fix: "+p.Fix, p.String())
			}

			s.Equal(tc.want, got)
		})
	}

	// Each check is a rule that Validate applies, and each rule has a
	// check: the cases above trigger every one.
	checked := make([]string, 0, len(endpoint.Checks))
	for _, c := range endpoint.Checks {
		checked = append(checked, c.Rule)
		s.NotEmpty(c.Requires, c.Rule)
	}
	triggered := make([]string, 0, len(seen))
	for r := range seen {
		triggered = append(triggered, r)
	}
	s.ElementsMatch(
		checked,
		triggered,
		"each check is triggered by a case, and each rule has a check",
	)
}

func TestEndpointPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(EndpointPublicTestSuite))
}

type badInIn struct {
	Q string `json:"q,omitempty" in:"qeury" validate:"omitempty" doc:"A filter"`
}

type secretInURLIn struct {
	ID    secret.Value `json:"id"              in:"path"  validate:"required"`
	Token secret.Value `json:"token,omitempty" in:"query" validate:"omitempty" telemetry:"attr"`
	Body  secret.Value `json:"body,omitempty"             validate:"omitempty"`
}

type exampleIn struct {
	Hostname string `json:"hostname"                  in:"path"  validate:"required"  flag:"target"   short:"T" doc:"The host"`
	Name     string `json:"hostname_filter,omitempty" in:"query" validate:"omitempty" flag:"hostname"           doc:"A filter"`
	Limit    int    `json:"limit,omitempty"           in:"query" validate:"omitempty"                           doc:"The page size"`
}

type exampleBodyIn struct {
	ID   string      `json:"id" in:"path" validate:"required"`
	Body exampleBody `          in:"body" validate:"required"`
}

type exampleBody struct {
	Reason string `json:"reason,omitempty" validate:"omitempty" doc:"Why"`
}
