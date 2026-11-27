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

package image

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	gohaidocker "github.com/osapi-io/gohai/pkg/gohai/collectors/docker"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// cli reads the images through gohai's docker collector, and changes them
// with the docker CLI, the same on each platform.
type cli struct {
	src  source.Sources
	exec host.Exec
}

// newCLI is the provider on each platform.
func newCLI(
	h host.Host,
) (Provider, error) {
	return &cli{src: h.Sources(), exec: h.Exec}, nil
}

// List reads the images. A host whose docker does not answer is skipped.
func (c *cli) List(
	ctx context.Context,
) ([]Entry, error) {
	out, err := gohaidocker.NewWith(c.src).Collect(ctx, nil)
	info, ok := out.(*gohaidocker.Info)
	if err != nil || !ok || info == nil {
		return nil, fmt.Errorf("docker does not answer on this host: %w",
			errors.Join(behavior.ErrUnsupported, err))
	}
	list := make([]Entry, 0, len(info.Images))
	for _, i := range info.Images {
		list = append(list, Entry{ID: i.ID, Repository: i.Repository, Tag: i.Tag, Size: i.Size})
	}
	slices.SortFunc(list, func(a, b Entry) int {
		return strings.Compare(a.Repository+":"+a.Tag, b.Repository+":"+b.Tag)
	})

	return list, nil
}

// inspect is the ID and the size of image on the host; an empty ID when the
// host does not have it. A failure of docker, such as a daemon that does not
// answer, is an error, not an image that the host does not have.
func (c *cli) inspect(
	ctx context.Context,
	image string,
) (string, int64, error) {
	out, err := c.exec.RunPrivilegedCmd(ctx, "docker",
		[]string{"image", "inspect", "--format", "{{.Id}} {{.Size}}", image})
	if err != nil {
		if strings.Contains(strings.ToLower(out), "no such image") {
			return "", 0, nil
		}

		return "", 0, fmt.Errorf("docker image inspect %s: %w: %s",
			image, err, strings.TrimSpace(out))
	}
	id, size, _ := strings.Cut(strings.TrimSpace(out), " ")
	n, _ := strconv.ParseInt(size, 10, 64)

	return id, n, nil
}

// Pull pulls the image, and reports whether the host got a new one.
func (c *cli) Pull(
	ctx context.Context,
	image string,
) (PullResult, error) {
	before, _, err := c.inspect(ctx, image)
	if err != nil {
		return PullResult{}, err
	}
	if out, err := c.exec.RunPrivilegedCmd(ctx, "docker", []string{"pull", image}); err != nil {
		return PullResult{}, fmt.Errorf(
			"docker pull %s: %w: %s",
			image,
			err,
			strings.TrimSpace(out),
		)
	}
	id, size, err := c.inspect(ctx, image)
	if err != nil {
		return PullResult{}, err
	}

	return PullResult{Image: image, ImageID: id, Size: size, Changed: id != before}, nil
}

// Delete removes the image. One that the host does not have does not change
// it.
func (c *cli) Delete(
	ctx context.Context,
	image string,
	force bool,
) (bool, error) {
	id, _, err := c.inspect(ctx, image)
	if err != nil {
		return false, err
	}
	if id == "" {
		return false, nil
	}
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	if out, err := c.exec.RunPrivilegedCmd(ctx, "docker", append(args, image)); err != nil {
		return false, fmt.Errorf("docker rmi %s: %w: %s", image, err, strings.TrimSpace(out))
	}

	return true, nil
}
