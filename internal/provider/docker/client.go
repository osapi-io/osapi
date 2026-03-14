// Copyright (c) 2025 John Dewey

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

package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// DockerAPIClient defines the subset of the Docker Engine API used by this
// provider.
type DockerAPIClient interface {
	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)
	ContainerStart(
		ctx context.Context,
		containerID string,
		options container.StartOptions,
	) error
	ContainerStop(
		ctx context.Context,
		containerID string,
		options container.StopOptions,
	) error
	ContainerRemove(
		ctx context.Context,
		containerID string,
		options container.RemoveOptions,
	) error
	ContainerList(
		ctx context.Context,
		options container.ListOptions,
	) ([]container.Summary, error)
	ContainerInspect(
		ctx context.Context,
		containerID string,
	) (container.InspectResponse, error)
	ContainerExecCreate(
		ctx context.Context,
		container string,
		options container.ExecOptions,
	) (common.IDResponse, error)
	ContainerExecAttach(
		ctx context.Context,
		execID string,
		options container.ExecStartOptions,
	) (types.HijackedResponse, error)
	ContainerExecInspect(
		ctx context.Context,
		execID string,
	) (container.ExecInspect, error)
	ImagePull(
		ctx context.Context,
		ref string,
		options image.PullOptions,
	) (io.ReadCloser, error)
	ImageInspect(
		ctx context.Context,
		imageID string,
		options ...dockerclient.ImageInspectOption,
	) (image.InspectResponse, error)
	ImageRemove(
		ctx context.Context,
		imageID string,
		options image.RemoveOptions,
	) ([]image.DeleteResponse, error)
	Ping(
		ctx context.Context,
	) (types.Ping, error)
}
