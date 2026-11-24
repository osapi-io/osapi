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

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/osapi-io/osapi/internal/cli"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	behaviorcommand "github.com/osapi-io/osapi/internal/behavior/command"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/behaviordsl"
	"github.com/osapi-io/osapi/internal/cli/commands/chain"
	"github.com/osapi-io/osapi/internal/cli/commands/clidocs"
	"github.com/osapi-io/osapi/internal/cli/commands/commanddsl"
	"github.com/osapi-io/osapi/internal/cli/commands/endpointdsl"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/telemetry/instruments"
)

// referenceDir is the root of the reference section of the docs site.
const referenceDir = "docs/docs/sidebar/reference/"

// referencePages are the pages of the reference: each is the output of the
// list or get command that shows it, so a page shows what its command shows.
func referencePages() ([]clidocs.Page, error) {
	ctx := context.Background()
	var pages []clidocs.Page

	bd := behaviorDSLDeps()
	bl, err := behaviordsl.List(ctx, bd, behaviordsl.ListInput{})
	if err != nil {
		return nil, err
	}
	pages = append(pages, clidocs.Page{
		Path:    referenceDir + "behaviors/_catalog.mdx",
		Partial: true,
		Parts: []clidocs.Part{
			{Group: "behavior", Out: bl, Link: column("NAME", behaviorPage)},
		},
	})
	pages = append(pages, behaviorFolders("behaviors/", func(*behavior.Behavior) bool {
		return true
	})...)
	for _, b := range behavior.All() {
		got, err := behaviordsl.Get(ctx, bd, behaviordsl.GetInput{Name: b.Name})
		if err != nil {
			return nil, err
		}
		// A behavior's page is in a folder of its own, in the folder of its
		// area, or of its parent when Of nests it: up is the path from there
		// to reference/behaviors.
		up := behaviorcommand.ReferenceUp(b)
		uses := columns(func(name string) string { return up + behaviorPage(name) },
			"Uses", "Used by")
		toBehavior := func(h string, cell string) string {
			switch {
			case h == "NAME" && b.Method(cell) != nil:
				return "#method-" + cell
			case h == "Command":
				return cliPageFrom(up+"../../", cell)
			}

			return uses(h, cell)
		}
		parts := []clidocs.Part{{Group: "behavior", Out: got, Link: toBehavior}}
		for _, m := range b.Methods {
			mg, err := behaviordsl.MethodGet(
				ctx,
				bd,
				behaviordsl.MethodInput{Name: b.Name, Method: m.Name},
			)
			if err != nil {
				return nil, err
			}
			parts = append(parts, clidocs.Part{
				Group: "behavior", Out: mg, Heading: "Method " + m.Name, Link: toBehavior,
			})
		}
		pages = append(pages, clidocs.Page{
			Path:  referenceDir + "behaviors/" + behaviorcommand.ReferencePage(b),
			Title: b.Title,
			Intro: b.Doc + ". `osapi dsl behavior get --name " + b.Name +
				"` and `osapi dsl behavior method get` show the same." + chainLink(b, up),
			Parts: parts,
		})
	}

	cd := chainDeps()
	cl, err := chain.List(ctx, cd, chain.ListInput{})
	if err != nil {
		return nil, err
	}
	pages = append(pages, clidocs.Page{
		Path:    referenceDir + "chains/_catalog.mdx",
		Partial: true,
		Parts: []clidocs.Part{
			{Group: "chain", Out: cl, Link: func(header string, cell string) string {
				switch header {
				case "BEHAVIOR":
					if b := behavior.Lookup(cell); b != nil && len(b.Requires) > 0 {
						return chainPage(b)
					}

					return "../behaviors/" + behaviorPage(cell)
				case "CALLS", "USED BY":
					if behavior.Lookup(cell) != nil {
						return "../behaviors/" + behaviorPage(cell)
					}
				}

				return ""
			}},
		},
	})
	pages = append(pages, behaviorFolders("chains/", func(b *behavior.Behavior) bool {
		return len(b.Requires) > 0
	})...)
	for _, b := range behavior.All() {
		if len(b.Requires) == 0 {
			continue
		}
		got, err := chain.Get(ctx, cd, chain.GetInput{Name: b.Name})
		if err != nil {
			return nil, err
		}
		pages = append(pages, clidocs.Page{
			Path:  referenceDir + "chains/" + chainPage(b),
			Title: b.Title,
			Intro: "The behaviors that " + b.Name + " calls, and each call that its examples make. " +
				"`osapi dsl chain get --name " + b.Name + "` shows the same.",
			Parts: []clidocs.Part{
				{Group: "chain", Out: got, Link: columns(func(name string) string {
					return "../../behaviors/" + behaviorPage(name)
				}, "Behavior", "Calls", "Used by")},
			},
		})
	}

	ed := endpointDSLDeps()
	el, err := endpointdsl.List(ctx, ed, endpointdsl.ListInput{})
	if err != nil {
		return nil, err
	}
	pages = append(pages, clidocs.Page{
		Path:    referenceDir + "endpoints/_catalog.mdx",
		Partial: true,
		Parts: []clidocs.Part{
			{Group: "endpoint", Out: el, Link: column("NAME", endpointPage)},
		},
	})
	pages = append(pages, endpointFolders()...)
	for _, g := range endpoint.All() {
		got, err := endpointdsl.Get(ctx, ed, endpointdsl.GetInput{Name: g.Name})
		if err != nil {
			return nil, err
		}
		pages = append(pages, clidocs.Page{
			Path:  referenceDir + "endpoints/" + endpointPage(g.Name),
			Title: endpointTitle(g),
			Intro: g.Doc + ". `osapi dsl endpoint get --name " + g.Name + "` shows the same.",
			Parts: []clidocs.Part{{Group: "endpoint", Out: got}},
		})
	}

	md := commandDSLDeps()
	ml, err := commanddsl.List(ctx, md, commanddsl.ListInput{})
	if err != nil {
		return nil, err
	}
	pages = append(pages, clidocs.Page{
		Path: referenceDir + "cli/_catalog.mdx", Partial: true,
		Parts: []clidocs.Part{{Group: "command", Out: ml, Link: column("COMMAND", cliPage)}},
	})
	for _, g := range command.All() {
		// A flat group at the root, such as version, has no command of its
		// own to page: its commands are in the CLI reference.
		if len(g.CommandPath()) == 0 {
			continue
		}
		// A flat group is found by its path, its words and its name, as
		// nats setup: two flat groups under nats have the same command
		// path, and so does the endpoint group nats. Its page has the name
		// of its command path when its name is the last word of it, as
		// agent, and the name of its path when not, as nats setup.
		words := g.CommandPath()
		if g.Flat && g.Name != words[len(words)-1] {
			words = append(words, g.Name)
		}
		name := strings.Join(words, " ")
		got, err := commanddsl.Get(ctx, md, commanddsl.GetInput{Name: g.Path()})
		if err != nil {
			return nil, err
		}
		pages = append(pages, clidocs.Page{
			Path:  referenceDir + "cli/" + strings.Join(words, "-") + ".md",
			Title: "osapi " + name,
			Intro: g.Doc + ". `osapi dsl command get --name " + g.Path() + "` shows the same.",
			Parts: []clidocs.Part{
				{Group: "command", Out: got, Link: func(h string, cell string) string {
					if h == cli.HeadingLink || h == "COMMAND" {
						return cliPage(cell)
					}

					return ""
				}},
			},
		})
	}

	return pages, nil
}

// column links each cell of the column header to the page that page names.
func column(
	header string,
	page func(cell string) string,
) func(string, string) string {
	return func(h string, cell string) string {
		if h != header {
			return ""
		}

		return page(cell)
	}
}

// cliPage is the page of the CLI reference of a command, such as osapi controller token
// generate, from a directory one level under the reference: a command with
// commands under it is the index page of its directory.
func cliPage(
	cell string,
) string {
	return cliPageFrom("../../", cell)
}

// cliPageFrom is the page of the CLI reference of a command, from the
// directory that up leads out of to the sidebar.
func cliPageFrom(
	up string,
	cell string,
) string {
	words := strings.Fields(strings.TrimPrefix(cell, "osapi "))
	if len(words) == 0 {
		return ""
	}
	c, _, err := rootCmd.Find(words)
	if err != nil || c == rootCmd {
		return ""
	}
	// The page is of the command that the tree finds, by its own path: a
	// heading such as agent status status names it with a word too many.
	words = strings.Fields(strings.TrimPrefix(c.CommandPath(), rootCmd.Name()+" "))
	if c.HasSubCommands() {
		return up + "usage/cli/" + strings.Join(words, "/") + "/" + words[len(words)-1] + ".mdx"
	}

	return up + "usage/cli/" + strings.Join(words, "/") + ".md"
}

// columns links each value of the columns or labels names to the page that
// page names, for a value that is a behavior.
func columns(
	page func(name string) string,
	names ...string,
) func(string, string) string {
	return func(h string, cell string) string {
		for _, n := range names {
			if strings.EqualFold(h, n) && behavior.Lookup(cell) != nil {
				return page(cell)
			}
		}

		return ""
	}
}

// chainLink is a sentence that links to the chain page of b, for a behavior
// that calls others, from the folder of b's page, which up leads out of.
func chainLink(
	b *behavior.Behavior,
	up string,
) string {
	if len(b.Requires) == 0 {
		return ""
	}

	return " [Its chain](" + up + "../chains/" + chainPage(b) +
		") shows each call that its examples make."
}

// chainPage is the chain page of b, from reference/chains: the page is in
// the folder of the area of b, the behavior at the top of the chain, such
// as config/sysctl.md.
func chainPage(
	b *behavior.Behavior,
) string {
	return behaviorcommand.ReferenceFolder(b) + "/" + b.Name + ".md"
}

// behaviorPage is the reference page of the behavior name, from
// reference/behaviors, such as hardware/disk/disk.md, or none for a name
// that no behavior has.
func behaviorPage(
	name string,
) string {
	b := behavior.Lookup(name)
	if b == nil {
		return ""
	}

	return behaviorcommand.ReferencePage(b)
}

// endpointPage is the reference page of the endpoint group name, from
// reference/endpoints: the page is in the folder of the group's area, such
// as fleet/agent.md.
func endpointPage(
	name string,
) string {
	for _, g := range endpoint.All() {
		if g.Name == name {
			return string(g.Area) + "/" + name + ".md"
		}
	}

	return ""
}

// behaviorFolders are the _category_.json of each folder under dir, of the
// reference, that has the page of a behavior that has: the folder of each
// area, and of the internal behaviors, last.
func behaviorFolders(
	dir string,
	has func(*behavior.Behavior) bool,
) []clidocs.Page {
	used := map[string]bool{}
	for _, b := range behavior.All() {
		if has(b) {
			used[behaviorcommand.ReferenceFolder(b)] = true
		}
	}
	folders := areaFolders(dir, used)
	if used[behaviorcommand.InternalFolder] {
		folders = append(folders, categoryPage(
			referenceDir+dir+behaviorcommand.InternalFolder,
			behaviorcommand.InternalLabel,
			len(area.All)+2,
		))
	}

	return folders
}

// endpointFolders are the _category_.json of each folder of an area that
// has an endpoint group's page.
func endpointFolders() []clidocs.Page {
	used := map[string]bool{}
	for _, g := range endpoint.All() {
		used[string(g.Area)] = true
	}

	return areaFolders("endpoints/", used)
}

// areaFolders are the _category_.json of each folder of an area under dir,
// of the reference, that used names, with the label and the order of
// area.All. A folder follows the index page of dir, the first page.
func areaFolders(
	dir string,
	used map[string]bool,
) []clidocs.Page {
	var folders []clidocs.Page
	for i, a := range area.All {
		if used[string(a.Name)] {
			folders = append(folders, categoryPage(referenceDir+dir+string(a.Name), a.Label, i+2))
		}
	}

	return folders
}

// categoryPage is the _category_.json of folder: its label and its place
// in the sidebar.
func categoryPage(
	folder string,
	label string,
	position int,
) clidocs.Page {
	return clidocs.Page{
		Path: folder + "/_category_.json",
		Text: fmt.Sprintf("{\n  \"label\": %q,\n  \"position\": %d\n}\n", label, position),
	}
}

// endpointInitialisms are the group names that a title keeps in capitals.
var endpointInitialisms = map[string]string{"nats": "NATS"}

// endpointTitle is the name of g as the title of its page.
func endpointTitle(
	g *endpoint.Group,
) string {
	if t, ok := endpointInitialisms[g.Name]; ok {
		return t
	}

	return strings.ToUpper(g.Name[:1]) + g.Name[1:]
}

// configTables is page with each generated table in it written again: the
// config tables, from config.Keys and the config that osapi ships under
// root, the role table, from authtoken.RolePermissions, the metrics
// reference, from instruments.All, and the screens of the CLI's output,
// from the examples of a behavior.
func configTables(
	root string,
	page string,
) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, config.ShippedFile))
	if err != nil {
		return "", err
	}
	shipped, err := config.Shipped(data)
	if err != nil {
		return "", err
	}

	page = config.DocTables(page, shipped)

	page = guide.Fill(page, "roles", func([]string) string {
		return authtoken.RoleTable()
	})

	page = guide.Fill(page, "metrics", func([]string) string {
		return instruments.Table()
	})

	return guide.Fill(page, "output", behaviorcommand.Screen), nil
}
