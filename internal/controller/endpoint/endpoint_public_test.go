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

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/apicontract"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

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
	Name  string   `json:"name"  validate:"required,oneof=a b"`
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
	doc, err := endpoint.OpenAPI(g, "../components.yaml#/components/schemas/ErrorResponse")
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

func (s *EndpointPublicTestSuite) TestFlagAndInternal() {
	m := endpoint.Get[noInput, Node]("get").Sub("trace").Flag("all", "Show everything").
		StringFlag("output", "", "The file to write", true)
	s.Equal([]endpoint.Flag{
		{Name: "all", Usage: "Show everything"},
		{Name: "output", Usage: "The file to write", String: true, Required: true},
	}, m.CommandFlags())
	s.Equal("trace", m.SubPath())
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
			s.Equal(tc.perm, g.Permission(tc.method))
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
					"body=application/json:{count:integer,name!:string(a|b),ratio:number,tags:[string]}! " +
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
}

func (s *EndpointPublicTestSuite) TestLocation() {
	m := endpoint.Post[noInput, Node]("make").Status(http.StatusCreated).Location("name")
	s.Equal("name", m.LocationField())

	doc, err := endpoint.OpenAPI(endpoint.Define("node", "The nodes", m), "x")
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
	doc, err := endpoint.OpenAPI(g, "x")
	s.Require().NoError(err)

	sample := doc.Components.Schemas["Sample"].Value
	s.Equal("A sample.", sample.Description)
	s.Equal(int64(4), sample.Properties["count"].Value.Example)
	s.InEpsilon(0.32, sample.Properties["load"].Value.Example, 0.0001)
	s.Equal(true, sample.Properties["on"].Value.Example)
	s.Equal(`{"host":"web-01"}`, sample.Properties["name"].Value.Example)
	s.Equal("many", sample.Properties["bad"].Value.Example)
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

// TestBodyValidateTags holds a body field to its whole validate tag: the
// items of an array, a pattern from startswith, and the tag in the extension.
// TestPrefix puts a group under a prefix whose {field} is a shared path
// parameter of each method, and keeps the method's own key after the name.
func (s *EndpointPublicTestSuite) TestPrefix() {
	g := endpoint.Define("sysctl", "Kernel parameters", endpoint.Prefix("/next/node/{hostname}/"),
		endpoint.Get[hostOnly, Node]("list"),
		endpoint.Get[hostKey, Node]("get"),
	)

	s.Equal("next/node/{hostname}", g.Prefix)
	s.Equal("/api/next/node/{hostname}/sysctl", g.Path(g.Method("list")))
	s.Equal("/api/next/node/{hostname}/sysctl/{key}", g.Path(g.Method("get")))
	s.Empty(g.Method("list").Key(), "the prefix's field is not the key")
	s.Equal("key", g.Method("get").Key())
	s.Empty(endpoint.Validate(g))

	doc, err := endpoint.OpenAPI(g, "x")
	s.Require().NoError(err)
	s.Equal([]string{"Hostname", "NodeKey"}, func() []string {
		names := make([]string, 0, len(doc.Components.Parameters))
		for name := range doc.Components.Parameters {
			names = append(names, name)
		}
		slices.Sort(names)

		return names
	}())
	get := doc.Paths.Value("/api/next/node/{hostname}/sysctl/{key}").Get
	s.Equal("#/components/parameters/Hostname", get.Parameters[0].Ref)
	s.Equal("#/components/parameters/NodeKey", get.Parameters[1].Ref)

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
		endpoint.CommandNotes("See the sysctl page."),
		endpoint.Get[noInput, Node]("list"),
		endpoint.Put[noInput, Node]("set").Permission("sysctl:admin"),
	)
	s.Equal("sysctl", g.GroupName())
	s.Equal("Kernel parameters", g.GroupDoc())
	s.Equal("See the sysctl page.", g.GroupNotes())

	methods := g.Commands()
	s.Require().Len(methods, 2)
	s.Equal("list", methods[0].MethodName())
	s.Equal(reflect.TypeFor[noInput](), methods[0].Input())
	s.Equal("sysctl:read", g.CommandPermission(methods[0]))
	s.Equal("sysctl:admin", g.CommandPermission(methods[1]))
}

func (s *EndpointPublicTestSuite) TestRefusals() {
	m := endpoint.Get[limited, Node]("get").Fails("a name over 3 characters", limited{Name: "abcd"})

	s.Equal([]endpoint.Refusal{
		{Doc: "a name over 3 characters", Input: limited{Name: "abcd"}},
	}, m.Refusals())
}

func (s *EndpointPublicTestSuite) TestBodyValidateTags() {
	doc, err := endpoint.OpenAPI(endpoint.Define("node", "The nodes",
		endpoint.Post[boundsInput, Node]("make")), "x")
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
		endpoint.Get[keyInput, Node]("get")), "x")
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
			_, err := endpoint.OpenAPI(endpoint.Define("node", "The nodes", tc.method), "x")

			s.ErrorContains(err, tc.wantErr)
		})
	}
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
				endpoint.Get[noInput, Node]("list"),
				endpoint.Get[noInput, Node]("get").Sub("one").Handle(func() {}),
			),
			want: []string{
				"the group Uses endpoint_test.deps, and the method has no handler.",
				"the handler is func(); it must be func(context.Context, endpoint_test.deps, endpoint_test.noInput) (endpoint_test.Node, error).",
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
			name:  "when the group has a bad name and no methods",
			group: endpoint.Define("Node", "The nodes"),
			want: []string{
				`group name "Node" is not lowercase letters and digits.`,
				"the group has no methods.",
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
				"the method has an action but no key.",
				"the input has more than one path field besides the prefix.",
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
