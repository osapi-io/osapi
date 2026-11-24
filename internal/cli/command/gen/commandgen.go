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

// Package commandgen writes the cobra commands of a group of the command
// layer, gen/cli.go: one command for each method, with its help, its flags
// and its examples, for the local generator, the endpoint generator and the
// behavior generator alike. A local group's command runs its method in the
// CLI with command.Run; an endpoint group's sends a request with the
// endpoint layer's Run.
package commandgen

import (
	"go/token"
	"reflect"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/secret"
)

const (
	// commandPath is the command layer, which registers the generated
	// commands and reads their files.
	commandPath = "github.com/osapi-io/osapi/internal/cli/command"
	// runPath is the endpoint layer's runner: its Run sends a command's
	// request. A local group's command calls the command layer's Run.
	runPath       = "github.com/osapi-io/osapi/internal/controller/endpoint/command"
	referencePath = "github.com/osapi-io/osapi/internal/cli/reference"
	cobraPath     = "github.com/spf13/cobra"
	configPath    = "github.com/osapi-io/osapi/internal/config"
)

// Commands writes gen/cli.go: one cobra command for each method, with its help,
// its permission, its examples and a flag for each input field, and the
// init that registers them under osapi client <group>, or under osapi for a
// local group.
func Commands(
	f *jen.File,
	g command.Group,
	pkg string,
) {
	_, isLocal := g.(*command.LocalGroup)
	f.ImportAlias(commandPath, "clicommand")
	if !isLocal {
		f.ImportAlias(runPath, "endpointcommand")
	}
	f.ImportName(referencePath, "reference")
	f.ImportName(cobraPath, "cobra")

	var inits []jen.Code
	methods := g.Commands()
	names := make([]jen.Code, 0, len(methods))
	for _, m := range methods {
		// The command's variable has each word of the method, such as
		// tokenRolesListCmd for roles list.
		name := g.GroupName()
		for _, w := range strings.Fields(m.MethodName()) {
			name += upperFirst(w)
		}
		name += "Cmd"
		names = append(names, jen.Id(name))

		var reads []jen.Code
		// flagFor reads fld from its flag, and declares the flag: named for
		// the field's json name with _ as -, or its flag tag, with the
		// shorthand of its short tag.
		flagFor := func(fld reflect.StructField) jen.Code {
			flag := strings.ReplaceAll(jsonName(fld), "_", "-")
			if f := fld.Tag.Get("flag"); f != "" {
				flag = f
			}
			kind, zero := flagKind(fld.Type, fld.Tag)
			local := lowerFirst(fld.Name)
			// A local that is a keyword, or the name of the group's package,
			// would hide it.
			if token.IsKeyword(local) || local == g.GroupName() {
				local += "Flag"
			}
			reads = append(reads, jen.List(jen.Id(local), jen.Id("_")).Op(":=").
				Id("cmd").Dot("Flags").Call().Dot("Get"+kind).Call(jen.Lit(flag)))
			// Flag help has no period at its end, as cobra's own has none.
			help := strings.TrimSuffix(fld.Tag.Get("doc"), ".")
			// A []byte is a file: its flag is the path of a local file, and
			// the command sends the file's content.
			if fld.Type == reflect.TypeFor[[]byte]() {
				path := local
				local += "Data"
				reads = append(reads,
					jen.List(jen.Id(local), jen.Id("ok")).Op(":=").
						Qual(commandPath, "ReadFile").Call(jen.Id(path)),
					jen.If(jen.Op("!").Id("ok")).Block(ret(isLocal)))
				help = "Path to the local file with " + strings.ToLower(help[:1]) + help[1:]
			}
			// A secret is a reference, or - for standard input: Secret and
			// Secrets refuse a literal value, which the shell history keeps.
			switch fld.Type {
			case reflect.TypeFor[secret.Value]():
				flagValue := local
				local += "Value"
				reads = append(reads,
					jen.List(jen.Id(local), jen.Id("ok")).Op(":=").
						Qual(commandPath, "Secret").
						Call(jen.Id("cmd"), jen.Lit(flag), jen.Id(flagValue)),
					jen.If(jen.Op("!").Id("ok")).Block(ret(isLocal)))
				help += command.SecretHelp
			case reflect.TypeFor[map[string]secret.Value]():
				flagValue := local
				local += "Value"
				reads = append(reads,
					jen.List(jen.Id(local), jen.Id("ok")).Op(":=").
						Qual(commandPath, "Secrets").
						Call(jen.Id("cmd"), jen.Lit(flag), jen.Id(flagValue)),
					jen.If(jen.Op("!").Id("ok")).Block(ret(isLocal)))
				help += command.SecretsHelp
			}
			if values := oneOf(fld); values != "" {
				help += " (" + values + ")"
			}
			usage := jen.Lit(help)
			declare := jen.Id(name).Dot("Flags").Call().Dot(kind).Call(jen.Lit(flag), zero, usage)
			if short := fld.Tag.Get("short"); short != "" {
				declare = jen.Id(name).Dot("Flags").Call().Dot(kind+"P").
					Call(jen.Lit(flag), jen.Lit(short), zero, usage)
			}
			inits = append(inits, declare)
			// A field with a default is never a required flag: the default
			// fills it, as --target is _all.
			_, hasDefault := fld.Tag.Lookup("default")
			if (fld.Tag.Get("in") == "path" || validates(fld, "required")) && !hasDefault {
				inits = append(inits, jen.Id("_").Op("=").Id(name).Dot("MarkFlagRequired").
					Call(jen.Lit(flag)))
			}

			return jen.Id(fld.Name).Op(":").Id(local)
		}
		var fields []jen.Code
		for _, fld := range structFields(m.Input()) {
			if fld.Tag.Get("in") != "body" || fld.Type.Kind() != reflect.Struct {
				fields = append(fields, flagFor(fld))

				continue
			}
			// A named body is a flag for each of its fields.
			var inner []jen.Code
			for _, bf := range structFields(fld.Type) {
				inner = append(inner, flagFor(bf))
			}
			fields = append(fields, jen.Id(fld.Name).Op(":").
				Qual(fld.Type.PkgPath(), fld.Type.Name()).Values(inner...))
		}
		in := jen.Qual(m.Input().PkgPath(), m.Input().Name())
		if len(fields) > 0 {
			in = in.Values(fields...)
		} else {
			in = in.Values()
		}
		// The flags are read first, then a blank line, then the call; a
		// method with no flags has only the call.
		run := reads
		if len(reads) > 0 {
			run = append(run, jen.Line())
		}
		call := jen.Qual(runPath, "Run").Call(
			jen.Id("cmd"), jen.Qual(pkg, "Group"), jen.Lit(m.MethodName()), in)
		if isLocal {
			call = jen.Return(jen.Qual(commandPath, "Run").Call(
				jen.Id("cmd"), jen.Qual(pkg, "Group"), jen.Lit(m.MethodName()), in))
		}
		// A process has no input: it runs until the command's context ends.
		if lm, ok := m.(*command.LocalMethod); ok && lm.IsProcess() {
			call = jen.Return(jen.Qual(commandPath, "Serve").Call(
				jen.Id("cmd"), jen.Qual(pkg, "Group"), jen.Lit(m.MethodName())))
			// Each key that Flag names is a flag that sets it.
			if keys := lm.ConfigKeys(); len(keys) > 0 {
				args := []jen.Code{jen.Id(name).Dot("Flags").Call()}
				for _, k := range keys {
					args = append(args, jen.Lit(k))
				}
				inits = append(inits, jen.Qual(configPath, "BindFlags").Call(args...))
			}
		}
		run = append(run, call)

		examples := make([]string, 0, len(m.Examples()))
		group := strings.Join(g.CommandPath(), " ")
		under := "osapi client "
		if isLocal {
			under = "osapi "
		}
		path := strings.TrimSpace(group + " " + m.CommandName())
		if sub := m.SubPath(); sub != "" && sub != m.CommandName() {
			path = strings.TrimSpace(group + " " + sub + " " + m.CommandName())
		}
		// An example with a doc has it as a # line above, and a blank line
		// between examples; examples with none are one to a line.
		sep := "\n"
		for _, e := range m.Examples() {
			line := strings.TrimRight("  "+under+path+" "+e.Flags, " ")
			if e.Doc != "" {
				line = "  # " + e.Doc + "\n" + line
				sep = "\n\n"
			}
			examples = append(examples, line)
		}

		dict := jen.Dict{
			jen.Id("Use"):   jen.Lit(m.CommandName()),
			jen.Id("Short"): jen.Lit(m.CommandShort()),
			jen.Id("Args"):  jen.Qual(cobraPath, "NoArgs"),
			jen.Id("Run"): jen.Func().Params(
				jen.Id("cmd").Op("*").Qual(cobraPath, "Command"),
				jen.Id("_").Index().String(),
			).Block(run...),
		}
		if isLocal {
			dict = jen.Dict{
				jen.Id("Use"):   jen.Lit(m.CommandName()),
				jen.Id("Short"): jen.Lit(m.CommandShort()),
				jen.Id("Args"):  jen.Qual(cobraPath, "NoArgs"),
				jen.Id("RunE"): jen.Func().Params(
					jen.Id("cmd").Op("*").Qual(cobraPath, "Command"),
					jen.Id("_").Index().String(),
				).Error().Block(run...),
			}
		}
		if long := m.CommandLong(); long != "" {
			dict[jen.Id("Long")] = jen.Lit(long)
		}
		if len(examples) > 0 {
			dict[jen.Id("Example")] = jen.Lit(strings.Join(examples, sep))
		}
		notes := jen.Dict{}
		if perm := g.CommandPermission(m); perm != "" {
			notes[jen.Qual(referencePath, "Permission")] = jen.Lit(string(perm))
		}
		if n := m.PageNotes(); n != "" {
			notes[jen.Qual(referencePath, "Notes")] = jen.Lit(n)
		}
		result := Outputs(g, m, strings.TrimSpace(under+path))
		if r := m.PageResult(); r != "" {
			result = strings.TrimSpace(result + "\n\n" + r)
		}
		if result != "" {
			notes[jen.Qual(referencePath, "Result")] = jen.Lit(result)
		}
		if sub := m.SubPath(); sub != "" && sub != m.CommandName() {
			notes[jen.Qual(commandPath, "SubKey")] = jen.Lit(sub)
		}
		if len(notes) > 0 {
			dict[jen.Id("Annotations")] = jen.Map(jen.String()).String().Values(notes)
		}
		f.Var().Id(name).Op("=").Op("&").Qual(cobraPath, "Command").Values(dict)
		f.Line()
	}

	inits = append(inits, jen.Qual(commandPath, "AddCommands").Call(
		append([]jen.Code{jen.Qual(pkg, "Group")}, names...)...))
	f.Func().Id("init").Params().Block(inits...)
}

// ret is the return of a generated command that stops early: a local
// command's RunE returns nil, as the failure is reported already.
func ret(
	local bool,
) jen.Code {
	if local {
		return jen.Return(jen.Nil())
	}

	return jen.Return()
}

// flagKind is the pflag kind of a field's type, such as String, and its
// value when the caller gives none: the field's default tag, or the zero
// value.
func flagKind(
	t reflect.Type,
	tag reflect.StructTag,
) (string, jen.Code) {
	d, ok := tag.Lookup("default")
	switch t.Kind() {
	case reflect.Map:
		// A map[string]string is key=value pairs, a flag given once each.
		return "StringToString", jen.Nil()
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			// A []string is a flag given once for each value, or with the
			// values joined with commas.
			return "StringSlice", jen.Nil()
		}

		return "String", jen.Lit(d)
	case reflect.Bool:
		return "Bool", jen.Lit(ok && d == "true")
	case reflect.Int:
		n, _ := strconv.Atoi(d)

		return "Int", jen.Lit(n)
	default:
		return "String", jen.Lit(d)
	}
}

func validates(
	f reflect.StructField,
	rule string,
) bool {
	for _, r := range strings.Split(f.Tag.Get("validate"), ",") {
		if r == rule {
			return true
		}
	}

	return false
}

func upperFirst(
	s string,
) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func lowerFirst(
	s string,
) string {
	// An initialism, such as ID, is lower case whole.
	if s == strings.ToUpper(s) {
		return strings.ToLower(s)
	}

	return strings.ToLower(s[:1]) + s[1:]
}

// oneOf is the values of a field's oneof rule, joined with ", ", or "".
func oneOf(
	f reflect.StructField,
) string {
	for _, r := range strings.Split(f.Tag.Get("validate"), ",") {
		if values, ok := strings.CutPrefix(r, "oneof="); ok {
			return strings.Join(strings.Fields(values), ", ")
		}
	}

	return ""
}

// structFields are the exported fields of t that JSON reads, when t is a
// struct: those with no json:"-".
func structFields(
	t reflect.Type,
) []reflect.StructField {
	if t.Kind() != reflect.Struct {
		return nil
	}
	var out []reflect.StructField
	for f := range t.Fields() {
		if f.IsExported() && jsonName(f) != "-" {
			out = append(out, f)
		}
	}

	return out
}

// jsonName is the name of f in JSON: its json tag's, or the field's.
func jsonName(
	f reflect.StructField,
) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		return f.Name
	}

	return name
}
