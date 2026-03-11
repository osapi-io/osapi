package docker_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/container/runtime"
	"github.com/retr0h/osapi/internal/provider/container/runtime/docker"
)

// mockDockerClient embeds dockerclient.APIClient and overrides specific methods for testing.
type mockDockerClient struct {
	dockerclient.APIClient
	pingFunc            func(ctx context.Context) (types.Ping, error)
	containerCreateFunc func(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)
	containerStartFunc      func(ctx context.Context, containerID string, options container.StartOptions) error
	containerStopFunc       func(ctx context.Context, containerID string, options container.StopOptions) error
	containerRemoveFunc     func(ctx context.Context, containerID string, options container.RemoveOptions) error
	containerListFunc       func(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	containerInspectFunc    func(ctx context.Context, containerID string) (types.ContainerJSON, error)
	containerExecCreateFunc func(
		ctx context.Context,
		containerID string,
		config container.ExecOptions,
	) (types.IDResponse, error)
	containerExecAttachFunc func(
		ctx context.Context,
		execID string,
		config container.ExecStartOptions,
	) (types.HijackedResponse, error)
	containerExecInspectFunc func(ctx context.Context, execID string) (container.ExecInspect, error)
	imagePullFunc            func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	imageInspectWithRawFunc  func(ctx context.Context, imageID string) (types.ImageInspect, []byte, error)
}

func (m *mockDockerClient) Ping(
	ctx context.Context,
) (types.Ping, error) {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}

	return types.Ping{}, nil
}

func (m *mockDockerClient) ContainerCreate(
	ctx context.Context,
	config *container.Config,
	hostConfig *container.HostConfig,
	networkConfig *network.NetworkingConfig,
	platform *ocispec.Platform,
	containerName string,
) (container.CreateResponse, error) {
	if m.containerCreateFunc != nil {
		return m.containerCreateFunc(ctx, config, hostConfig, networkConfig, platform, containerName)
	}

	return container.CreateResponse{ID: "test-container-id"}, nil
}

func (m *mockDockerClient) ContainerStart(
	ctx context.Context,
	containerID string,
	options container.StartOptions,
) error {
	if m.containerStartFunc != nil {
		return m.containerStartFunc(ctx, containerID, options)
	}

	return nil
}

func (m *mockDockerClient) ContainerStop(
	ctx context.Context,
	containerID string,
	options container.StopOptions,
) error {
	if m.containerStopFunc != nil {
		return m.containerStopFunc(ctx, containerID, options)
	}

	return nil
}

func (m *mockDockerClient) ContainerRemove(
	ctx context.Context,
	containerID string,
	options container.RemoveOptions,
) error {
	if m.containerRemoveFunc != nil {
		return m.containerRemoveFunc(ctx, containerID, options)
	}

	return nil
}

func (m *mockDockerClient) ContainerList(
	ctx context.Context,
	options container.ListOptions,
) ([]types.Container, error) {
	if m.containerListFunc != nil {
		return m.containerListFunc(ctx, options)
	}

	return []types.Container{}, nil
}

func (m *mockDockerClient) ContainerInspect(
	ctx context.Context,
	containerID string,
) (types.ContainerJSON, error) {
	if m.containerInspectFunc != nil {
		return m.containerInspectFunc(ctx, containerID)
	}

	return types.ContainerJSON{}, nil
}

func (m *mockDockerClient) ContainerExecCreate(
	ctx context.Context,
	containerID string,
	config container.ExecOptions,
) (types.IDResponse, error) {
	if m.containerExecCreateFunc != nil {
		return m.containerExecCreateFunc(ctx, containerID, config)
	}

	return types.IDResponse{ID: "test-exec-id"}, nil
}

func (m *mockDockerClient) ContainerExecAttach(
	ctx context.Context,
	execID string,
	config container.ExecStartOptions,
) (types.HijackedResponse, error) {
	if m.containerExecAttachFunc != nil {
		return m.containerExecAttachFunc(ctx, execID, config)
	}

	return types.HijackedResponse{}, nil
}

func (m *mockDockerClient) ContainerExecInspect(
	ctx context.Context,
	execID string,
) (container.ExecInspect, error) {
	if m.containerExecInspectFunc != nil {
		return m.containerExecInspectFunc(ctx, execID)
	}

	return container.ExecInspect{}, nil
}

func (m *mockDockerClient) ImagePull(
	ctx context.Context,
	ref string,
	options image.PullOptions,
) (io.ReadCloser, error) {
	if m.imagePullFunc != nil {
		return m.imagePullFunc(ctx, ref, options)
	}

	return io.NopCloser(strings.NewReader("{}")), nil
}

func (m *mockDockerClient) ImageInspectWithRaw(
	ctx context.Context,
	imageID string,
) (types.ImageInspect, []byte, error) {
	if m.imageInspectWithRawFunc != nil {
		return m.imageInspectWithRawFunc(ctx, imageID)
	}

	return types.ImageInspect{
		ID:       "sha256:test",
		RepoTags: []string{"nginx:latest"},
		Size:     1024,
	}, nil, nil
}

type DockerDriverPublicTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *DockerDriverPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *DockerDriverPublicTestSuite) TestNew() {
	tests := []struct {
		name         string
		validateFunc func(d runtime.Driver, err error)
	}{
		{
			name: "returns non-nil driver",
			validateFunc: func(
				d runtime.Driver,
				err error,
			) {
				s.NoError(err)
				s.NotNil(d)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d, err := docker.New()
			tt.validateFunc(d, err)
		})
	}
}

func (s *DockerDriverPublicTestSuite) TestPing() {
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		validateFunc func(err error)
	}{
		{
			name: "successful ping",
			mockClient: &mockDockerClient{
				pingFunc: func(
					ctx context.Context,
				) (types.Ping, error) {
					return types.Ping{APIVersion: "1.45"}, nil
				},
			},
			validateFunc: func(
				err error,
			) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			err := d.Ping(s.ctx)
			tt.validateFunc(err)
		})
	}
}

func (s *DockerDriverPublicTestSuite) TestCreate() {
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		params       runtime.CreateParams
		validateFunc func(c *runtime.Container, err error)
	}{
		{
			name: "successful container creation",
			mockClient: &mockDockerClient{
				containerCreateFunc: func(
					ctx context.Context,
					config *container.Config,
					hostConfig *container.HostConfig,
					networkConfig *network.NetworkingConfig,
					platform *ocispec.Platform,
					containerName string,
				) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "test-id"}, nil
				},
			},
			params: runtime.CreateParams{
				Image: "nginx:latest",
				Name:  "test-nginx",
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.NotNil(c)
				s.Equal("test-id", c.ID)
			},
		},
		{
			name: "successful container creation with auto-start",
			mockClient: &mockDockerClient{
				containerCreateFunc: func(
					ctx context.Context,
					config *container.Config,
					hostConfig *container.HostConfig,
					networkConfig *network.NetworkingConfig,
					platform *ocispec.Platform,
					containerName string,
				) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "test-id-auto"}, nil
				},
				containerStartFunc: func(
					ctx context.Context,
					containerID string,
					options container.StartOptions,
				) error {
					return nil
				},
			},
			params: runtime.CreateParams{
				Image:     "nginx:latest",
				Name:      "test-nginx-auto",
				AutoStart: true,
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.NotNil(c)
				s.Equal("test-id-auto", c.ID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			c, err := d.Create(s.ctx, tt.params)
			tt.validateFunc(c, err)
		})
	}
}

func (s *DockerDriverPublicTestSuite) TestStart() {
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		containerID  string
		validateFunc func(err error)
	}{
		{
			name: "successful start",
			mockClient: &mockDockerClient{
				containerStartFunc: func(
					ctx context.Context,
					containerID string,
					options container.StartOptions,
				) error {
					return nil
				},
			},
			containerID: "test-id",
			validateFunc: func(
				err error,
			) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			err := d.Start(s.ctx, tt.containerID)
			tt.validateFunc(err)
		})
	}
}

func (s *DockerDriverPublicTestSuite) TestStop() {
	timeout := 10 * time.Second
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		containerID  string
		timeout      *time.Duration
		validateFunc func(err error)
	}{
		{
			name: "successful stop with timeout",
			mockClient: &mockDockerClient{
				containerStopFunc: func(
					ctx context.Context,
					containerID string,
					options container.StopOptions,
				) error {
					return nil
				},
			},
			containerID: "test-id",
			timeout:     &timeout,
			validateFunc: func(
				err error,
			) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			err := d.Stop(s.ctx, tt.containerID, tt.timeout)
			tt.validateFunc(err)
		})
	}
}

func (s *DockerDriverPublicTestSuite) TestRemove() {
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		containerID  string
		force        bool
		validateFunc func(err error)
	}{
		{
			name: "successful remove with force",
			mockClient: &mockDockerClient{
				containerRemoveFunc: func(
					ctx context.Context,
					containerID string,
					options container.RemoveOptions,
				) error {
					return nil
				},
			},
			containerID: "test-id",
			force:       true,
			validateFunc: func(
				err error,
			) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			err := d.Remove(s.ctx, tt.containerID, tt.force)
			tt.validateFunc(err)
		})
	}
}

func (s *DockerDriverPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		params       runtime.ListParams
		validateFunc func(containers []runtime.Container, err error)
	}{
		{
			name: "successful list all containers",
			mockClient: &mockDockerClient{
				containerListFunc: func(
					ctx context.Context,
					options container.ListOptions,
				) ([]types.Container, error) {
					return []types.Container{
						{
							ID:      "test-id-1",
							Names:   []string{"/test-container-1"},
							Image:   "nginx:latest",
							State:   "running",
							Created: time.Now().Unix(),
						},
					}, nil
				},
			},
			params: runtime.ListParams{State: "all"},
			validateFunc: func(
				containers []runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.Len(containers, 1)
				s.Equal("test-id-1", containers[0].ID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			containers, err := d.List(s.ctx, tt.params)
			tt.validateFunc(containers, err)
		})
	}
}

func (s *DockerDriverPublicTestSuite) TestInspect() {
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		containerID  string
		validateFunc func(detail *runtime.ContainerDetail, err error)
	}{
		{
			name: "successful inspect",
			mockClient: &mockDockerClient{
				containerInspectFunc: func(
					ctx context.Context,
					containerID string,
				) (types.ContainerJSON, error) {
					return types.ContainerJSON{
						ContainerJSONBase: &types.ContainerJSONBase{
							ID:      "test-id",
							Name:    "/test-container",
							State:   &types.ContainerState{Status: "running"},
							Created: time.Now().Format(time.RFC3339Nano),
						},
						Config: &container.Config{Image: "nginx:latest"},
					}, nil
				},
			},
			containerID: "test-id",
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.NoError(err)
				s.NotNil(detail)
				s.Equal("test-id", detail.ID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			detail, err := d.Inspect(s.ctx, tt.containerID)
			tt.validateFunc(detail, err)
		})
	}
}

func (s *DockerDriverPublicTestSuite) TestPull() {
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		imageName    string
		validateFunc func(result *runtime.PullResult, err error)
	}{
		{
			name: "successful image pull",
			mockClient: &mockDockerClient{
				imagePullFunc: func(
					ctx context.Context,
					ref string,
					options image.PullOptions,
				) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("{}")), nil
				},
				imageInspectWithRawFunc: func(
					ctx context.Context,
					imageID string,
				) (types.ImageInspect, []byte, error) {
					return types.ImageInspect{
						ID:       "sha256:test123",
						RepoTags: []string{"nginx:latest"},
						Size:     2048,
					}, nil, nil
				},
			},
			imageName: "nginx:latest",
			validateFunc: func(
				result *runtime.PullResult,
				err error,
			) {
				s.NoError(err)
				s.NotNil(result)
				s.Equal("sha256:test123", result.ImageID)
				s.Equal("latest", result.Tag)
				s.Equal(int64(2048), result.Size)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			result, err := d.Pull(s.ctx, tt.imageName)
			tt.validateFunc(result, err)
		})
	}
}

func TestDockerDriverPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DockerDriverPublicTestSuite))
}
