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

// Package image is the image behavior: the Docker images of the host. It
// reads them through gohai's docker collector, and pulls and removes them
// with the docker CLI, on each platform where docker answers.
package image

import (
	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
)

// The examples' image. The integration test runs the write examples on a
// live host, each from its image removed first: alpine is small.
const (
	image = "alpine:3.20"
	id    = "sha256:91ef0af61f39ece4d6710e465df5ed6ca12112358344fd51ae6a3b886634148b"
)

// The commands of the examples' host.
const (
	versionCmd = "docker version --format {{.Server.Version}}"
	imagesCmd  = "docker images --format {{json .}}"
	inspectCmd = "docker image inspect --format {{.Id}} {{.Size}} alpine:3.20"
)

var (
	answers = behavior.Prints(versionCmd, "29.1.3\n")
	pulled  = behavior.PrintsAfter("docker pull alpine:3.20", inspectCmd, id+" 3634816\n")
	gone    = behavior.PrintsAfter("docker rmi alpine:3.20", inspectCmd, "")
)

// Behavior is the image definition.
var Behavior = behavior.Register(behavior.Define(
	"image", "The Docker images of the host",
	behavior.Title("Container image"),
	behavior.Node,
	behavior.API,
	behavior.Area(area.Containers),
	behavior.CLI,
	behavior.Platform(behavior.PlatformLinux, newCLI),
	behavior.Platform(behavior.PlatformDarwin, newCLI),
	behavior.Sudo("docker"),
	behavior.Runs("docker"),
	behavior.Read("list", List).
		Describe("List the images of the host").
		Explain("Reads each image through gohai's docker collector, which runs docker images. "+
			"A host whose docker does not answer reports skipped.").
		Example("A host with an image", ListInput{},
			[]Entry{{ID: "91ef0af61f39", Repository: "alpine", Tag: "3.20", Size: "7.8MB"}},
			answers,
			behavior.Prints(imagesCmd,
				`{"ID":"91ef0af61f39","Repository":"alpine","Tag":"3.20","Size":"7.8MB"}`+"\n"),
		),
	behavior.Action("pull", Pull).
		Describe("Pull an image").
		Explain("Runs docker pull, and answers the ID and the size of the image that the host "+
			"has after it. A pull that gets nothing new, because the host had the image of "+
			"that reference already, reports changed: false.").
		Example("Pull an image", PullInput{Image: image},
			PullResult{Image: image, ImageID: id, Size: 3634816, Changed: true},
			pulled,
			behavior.Ran("docker", "pull", image),
		).
		Fails("An image that is an option", PullInput{Image: "--all-tags"}, "image_ref").
		Fails("No image", PullInput{}, "required"),
	behavior.Write("delete", Delete).Key("image").
		Describe("Remove an image").
		Explain("Runs docker rmi, with -f when force is set, which removes an image that a "+
			"container uses. An image that the host does not have reports changed: false.").
		Example("Remove an image", DeleteInput{Image: image}, Result{Image: image, Changed: true},
			pulled, gone,
			behavior.After(PullInput{Image: image}),
			behavior.Ran("docker", "rmi", image),
		).
		Example("An image that the host does not have", DeleteInput{Image: image}, Result{Image: image}).
		Fails("An image that is an option", DeleteInput{Image: "-f"}, "image_ref"),
))
