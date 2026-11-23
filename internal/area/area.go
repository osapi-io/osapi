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

// Package area is the areas that group the operations of osapi: the tabs of
// the UI's Configure page, the folders of the SDK's pages and of the
// reference pages of the behaviors, the chains and the endpoints, and the
// sections of the docs site's SDK and Behaviors menus. Each behavior and each endpoint
// group declares one. The spec carries the area of each operation and the
// list of areas, so a change here changes each of them together.
package area

import "strings"

// Name is the name of an area, one of All.
type Name string

// The areas.
const (
	Services    Name = "services"
	Software    Name = "software"
	Config      Name = "config"
	Node        Name = "node"
	Logs        Name = "logs"
	Networking  Name = "networking"
	Security    Name = "security"
	Containers  Name = "containers"
	ObjectStore Name = "objectstore"
	Command     Name = "command"
	Hardware    Name = "hardware"
	Audit       Name = "audit"
	Fleet       Name = "fleet"
)

// Area is one area: its name, and the label that a reader sees.
type Area struct {
	Name  Name   `json:"name"`
	Label string `json:"label"`
}

// All are the areas, in the order that the UI, the SDK's pages and the docs
// site show them.
var All = []Area{
	{Services, "Services"},
	{Software, "Software"},
	{Config, "Config"},
	{Node, "Node"},
	{Logs, "Logs"},
	{Networking, "Networking"},
	{Security, "Security"},
	{Containers, "Containers"},
	{ObjectStore, "Object Store"},
	{Command, "Command"},
	{Hardware, "Hardware"},
	{Audit, "Audit"},
	{Fleet, "Fleet"},
}

// Valid reports whether name is one of All.
func Valid(
	name Name,
) bool {
	for _, a := range All {
		if a.Name == name {
			return true
		}
	}

	return false
}

// List is the names of All, for a message, such as "services, software".
func List() string {
	names := make([]string, 0, len(All))
	for _, a := range All {
		names = append(names, string(a.Name))
	}

	return strings.Join(names, ", ")
}
