package docker_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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
	containerListFunc       func(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	containerInspectFunc    func(ctx context.Context, containerID string) (container.InspectResponse, error)
	containerExecCreateFunc func(
		ctx context.Context,
		containerID string,
		config container.ExecOptions,
	) (container.ExecCreateResponse, error)
	containerExecAttachFunc func(
		ctx context.Context,
		execID string,
		config container.ExecStartOptions,
	) (types.HijackedResponse, error)
	containerExecInspectFunc func(ctx context.Context, execID string) (container.ExecInspect, error)
	imagePullFunc            func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	imageInspectFunc         func(ctx context.Context, imageID string, options ...dockerclient.ImageInspectOption) (image.InspectResponse, error)
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
		return m.containerCreateFunc(
			ctx,
			config,
			hostConfig,
			networkConfig,
			platform,
			containerName,
		)
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
) ([]container.Summary, error) {
	if m.containerListFunc != nil {
		return m.containerListFunc(ctx, options)
	}

	return []container.Summary{}, nil
}

func (m *mockDockerClient) ContainerInspect(
	ctx context.Context,
	containerID string,
) (container.InspectResponse, error) {
	if m.containerInspectFunc != nil {
		return m.containerInspectFunc(ctx, containerID)
	}

	return container.InspectResponse{}, nil
}

func (m *mockDockerClient) ContainerExecCreate(
	ctx context.Context,
	containerID string,
	config container.ExecOptions,
) (container.ExecCreateResponse, error) {
	if m.containerExecCreateFunc != nil {
		return m.containerExecCreateFunc(ctx, containerID, config)
	}

	return container.ExecCreateResponse{ID: "test-exec-id"}, nil
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

func (m *mockDockerClient) ImageInspect(
	ctx context.Context,
	imageID string,
	options ...dockerclient.ImageInspectOption,
) (image.InspectResponse, error) {
	if m.imageInspectFunc != nil {
		return m.imageInspectFunc(ctx, imageID, options...)
	}

	return image.InspectResponse{
		ID:       "sha256:test",
		RepoTags: []string{"nginx:latest"},
		Size:     1024,
	}, nil
}

// newHijackedResponse creates a HijackedResponse with the given content
// suitable for testing. It uses a pipe-based net.Conn so Close() does not panic.
func newHijackedResponse(
	content string,
) types.HijackedResponse {
	serverConn, clientConn := net.Pipe()
	_ = serverConn.Close()

	return types.HijackedResponse{
		Conn:   clientConn,
		Reader: bufio.NewReader(strings.NewReader(content)),
	}
}

// errorConn is a net.Conn that always returns an error on Read.
type errorConn struct {
	net.Conn
}

func (c *errorConn) Read(
	_ []byte,
) (int, error) {
	return 0, errors.New("forced read error")
}

func (c *errorConn) Close() error {
	return nil
}

func newErrorHijackedResponse() types.HijackedResponse {
	return types.HijackedResponse{
		Conn:   &errorConn{},
		Reader: bufio.NewReader(&errorConn{}),
	}
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
		setupEnv     func() (cleanup func())
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
		{
			name: "returns error when docker client creation fails",
			setupEnv: func() (cleanup func()) {
				orig, existed := os.LookupEnv("DOCKER_HOST")
				_ = os.Setenv("DOCKER_HOST", "tcp://invalid:::::port")

				return func() {
					if existed {
						_ = os.Setenv("DOCKER_HOST", orig)
					} else {
						_ = os.Unsetenv("DOCKER_HOST")
					}
				}
			},
			validateFunc: func(
				d runtime.Driver,
				err error,
			) {
				s.Error(err)
				s.Contains(err.Error(), "create docker client")
				s.Nil(d)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setupEnv != nil {
				cleanup := tt.setupEnv()
				defer cleanup()
			}

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
					_ context.Context,
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
		{
			name: "returns error when ping fails",
			mockClient: &mockDockerClient{
				pingFunc: func(
					_ context.Context,
				) (types.Ping, error) {
					return types.Ping{}, fmt.Errorf("connection refused")
				},
			},
			validateFunc: func(
				err error,
			) {
				s.Error(err)
				s.Contains(err.Error(), "ping docker daemon")
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
					_ context.Context,
					_ *container.Config,
					_ *container.HostConfig,
					_ *network.NetworkingConfig,
					_ *ocispec.Platform,
					_ string,
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
					_ context.Context,
					_ *container.Config,
					_ *container.HostConfig,
					_ *network.NetworkingConfig,
					_ *ocispec.Platform,
					_ string,
				) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "test-id-auto"}, nil
				},
				containerStartFunc: func(
					_ context.Context,
					_ string,
					_ container.StartOptions,
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
		{
			name: "with command set",
			mockClient: &mockDockerClient{
				containerCreateFunc: func(
					_ context.Context,
					config *container.Config,
					_ *container.HostConfig,
					_ *network.NetworkingConfig,
					_ *ocispec.Platform,
					_ string,
				) (container.CreateResponse, error) {
					s.Equal([]string(config.Cmd), []string{"echo", "hello"})

					return container.CreateResponse{ID: "cmd-id"}, nil
				},
			},
			params: runtime.CreateParams{
				Image:   "alpine:latest",
				Name:    "test-cmd",
				Command: []string{"echo", "hello"},
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.NotNil(c)
				s.Equal("cmd-id", c.ID)
			},
		},
		{
			name: "with env map",
			mockClient: &mockDockerClient{
				containerCreateFunc: func(
					_ context.Context,
					config *container.Config,
					_ *container.HostConfig,
					_ *network.NetworkingConfig,
					_ *ocispec.Platform,
					_ string,
				) (container.CreateResponse, error) {
					s.Len(config.Env, 1)
					s.Contains(config.Env[0], "FOO=bar")

					return container.CreateResponse{ID: "env-id"}, nil
				},
			},
			params: runtime.CreateParams{
				Image: "alpine:latest",
				Name:  "test-env",
				Env:   map[string]string{"FOO": "bar"},
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.NotNil(c)
				s.Equal("env-id", c.ID)
			},
		},
		{
			name: "with ports",
			mockClient: &mockDockerClient{
				containerCreateFunc: func(
					_ context.Context,
					config *container.Config,
					hostConfig *container.HostConfig,
					_ *network.NetworkingConfig,
					_ *ocispec.Platform,
					_ string,
				) (container.CreateResponse, error) {
					s.NotNil(config.ExposedPorts)
					s.NotNil(hostConfig.PortBindings)
					containerPort := nat.Port("80/tcp")
					s.Contains(hostConfig.PortBindings, containerPort)
					s.Equal("8080", hostConfig.PortBindings[containerPort][0].HostPort)

					return container.CreateResponse{ID: "ports-id"}, nil
				},
			},
			params: runtime.CreateParams{
				Image: "nginx:latest",
				Name:  "test-ports",
				Ports: []runtime.PortMapping{
					{Host: 8080, Container: 80},
				},
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.NotNil(c)
				s.Equal("ports-id", c.ID)
			},
		},
		{
			name: "with volumes",
			mockClient: &mockDockerClient{
				containerCreateFunc: func(
					_ context.Context,
					_ *container.Config,
					hostConfig *container.HostConfig,
					_ *network.NetworkingConfig,
					_ *ocispec.Platform,
					_ string,
				) (container.CreateResponse, error) {
					s.Len(hostConfig.Mounts, 1)
					s.Equal("/host/data", hostConfig.Mounts[0].Source)
					s.Equal("/container/data", hostConfig.Mounts[0].Target)

					return container.CreateResponse{ID: "vols-id"}, nil
				},
			},
			params: runtime.CreateParams{
				Image: "nginx:latest",
				Name:  "test-vols",
				Volumes: []runtime.VolumeMapping{
					{Host: "/host/data", Container: "/container/data"},
				},
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.NotNil(c)
				s.Equal("vols-id", c.ID)
			},
		},
		{
			name: "returns error when container create fails",
			mockClient: &mockDockerClient{
				containerCreateFunc: func(
					_ context.Context,
					_ *container.Config,
					_ *container.HostConfig,
					_ *network.NetworkingConfig,
					_ *ocispec.Platform,
					_ string,
				) (container.CreateResponse, error) {
					return container.CreateResponse{}, fmt.Errorf("image not found")
				},
			},
			params: runtime.CreateParams{
				Image: "nonexistent:latest",
				Name:  "test-fail",
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.Error(err)
				s.Nil(c)
				s.Contains(err.Error(), "create container")
			},
		},
		{
			name: "returns error when auto-start fails",
			mockClient: &mockDockerClient{
				containerCreateFunc: func(
					_ context.Context,
					_ *container.Config,
					_ *container.HostConfig,
					_ *network.NetworkingConfig,
					_ *ocispec.Platform,
					_ string,
				) (container.CreateResponse, error) {
					return container.CreateResponse{ID: "autostart-fail-id"}, nil
				},
				containerStartFunc: func(
					_ context.Context,
					_ string,
					_ container.StartOptions,
				) error {
					return fmt.Errorf("start failed")
				},
			},
			params: runtime.CreateParams{
				Image:     "nginx:latest",
				Name:      "test-autostart-fail",
				AutoStart: true,
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.Error(err)
				s.Nil(c)
				s.Contains(err.Error(), "auto-start container")
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
					_ context.Context,
					_ string,
					_ container.StartOptions,
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
		{
			name: "returns error when start fails",
			mockClient: &mockDockerClient{
				containerStartFunc: func(
					_ context.Context,
					_ string,
					_ container.StartOptions,
				) error {
					return fmt.Errorf("container not found")
				},
			},
			containerID: "missing-id",
			validateFunc: func(
				err error,
			) {
				s.Error(err)
				s.Contains(err.Error(), "start container")
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
					_ context.Context,
					_ string,
					options container.StopOptions,
				) error {
					s.NotNil(options.Timeout)

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
		{
			name: "successful stop with nil timeout",
			mockClient: &mockDockerClient{
				containerStopFunc: func(
					_ context.Context,
					_ string,
					options container.StopOptions,
				) error {
					s.Nil(options.Timeout)

					return nil
				},
			},
			containerID: "test-id",
			timeout:     nil,
			validateFunc: func(
				err error,
			) {
				s.NoError(err)
			},
		},
		{
			name: "returns error when stop fails",
			mockClient: &mockDockerClient{
				containerStopFunc: func(
					_ context.Context,
					_ string,
					_ container.StopOptions,
				) error {
					return fmt.Errorf("container not running")
				},
			},
			containerID: "stopped-id",
			timeout:     nil,
			validateFunc: func(
				err error,
			) {
				s.Error(err)
				s.Contains(err.Error(), "stop container")
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
					_ context.Context,
					_ string,
					_ container.RemoveOptions,
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
		{
			name: "returns error when remove fails",
			mockClient: &mockDockerClient{
				containerRemoveFunc: func(
					_ context.Context,
					_ string,
					_ container.RemoveOptions,
				) error {
					return fmt.Errorf("container is running")
				},
			},
			containerID: "running-id",
			force:       false,
			validateFunc: func(
				err error,
			) {
				s.Error(err)
				s.Contains(err.Error(), "remove container")
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
					_ context.Context,
					options container.ListOptions,
				) ([]container.Summary, error) {
					s.True(options.All)

					return []container.Summary{
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
		{
			name: "state running sets all to false",
			mockClient: &mockDockerClient{
				containerListFunc: func(
					_ context.Context,
					options container.ListOptions,
				) ([]container.Summary, error) {
					s.False(options.All)

					return []container.Summary{}, nil
				},
			},
			params: runtime.ListParams{State: "running"},
			validateFunc: func(
				containers []runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.Empty(containers)
			},
		},
		{
			name: "state stopped sets filter",
			mockClient: &mockDockerClient{
				containerListFunc: func(
					_ context.Context,
					options container.ListOptions,
				) ([]container.Summary, error) {
					s.True(options.All)
					s.NotNil(options.Filters)

					return []container.Summary{}, nil
				},
			},
			params: runtime.ListParams{State: "stopped"},
			validateFunc: func(
				containers []runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.Empty(containers)
			},
		},
		{
			name: "default empty state",
			mockClient: &mockDockerClient{
				containerListFunc: func(
					_ context.Context,
					options container.ListOptions,
				) ([]container.Summary, error) {
					s.False(options.All)

					return []container.Summary{}, nil
				},
			},
			params: runtime.ListParams{State: ""},
			validateFunc: func(
				containers []runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.Empty(containers)
			},
		},
		{
			name: "with limit",
			mockClient: &mockDockerClient{
				containerListFunc: func(
					_ context.Context,
					options container.ListOptions,
				) ([]container.Summary, error) {
					s.Equal(5, options.Limit)

					return []container.Summary{}, nil
				},
			},
			params: runtime.ListParams{State: "all", Limit: 5},
			validateFunc: func(
				containers []runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.Empty(containers)
			},
		},
		{
			name: "container with empty names slice",
			mockClient: &mockDockerClient{
				containerListFunc: func(
					_ context.Context,
					_ container.ListOptions,
				) ([]container.Summary, error) {
					return []container.Summary{
						{
							ID:      "no-name-id",
							Names:   []string{},
							Image:   "alpine:latest",
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
				s.Equal("", containers[0].Name)
			},
		},
		{
			name: "returns error when list fails",
			mockClient: &mockDockerClient{
				containerListFunc: func(
					_ context.Context,
					_ container.ListOptions,
				) ([]container.Summary, error) {
					return nil, fmt.Errorf("daemon error")
				},
			},
			params: runtime.ListParams{State: "all"},
			validateFunc: func(
				containers []runtime.Container,
				err error,
			) {
				s.Error(err)
				s.Nil(containers)
				s.Contains(err.Error(), "list containers")
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
					_ context.Context,
					_ string,
				) (container.InspectResponse, error) {
					return container.InspectResponse{
						ContainerJSONBase: &container.ContainerJSONBase{
							ID:      "test-id",
							Name:    "/test-container",
							State:   &container.State{Status: "running"},
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
				s.Equal("running", detail.State)
			},
		},
		{
			name: "nil state returns unknown",
			mockClient: &mockDockerClient{
				containerInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.InspectResponse, error) {
					return container.InspectResponse{
						ContainerJSONBase: &container.ContainerJSONBase{
							ID:      "nil-state-id",
							Name:    "/nil-state",
							State:   nil,
							Created: time.Now().Format(time.RFC3339Nano),
						},
						Config: &container.Config{Image: "nginx:latest"},
					}, nil
				},
			},
			containerID: "nil-state-id",
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.NoError(err)
				s.NotNil(detail)
				s.Equal("unknown", detail.State)
			},
		},
		{
			name: "invalid created timestamp falls back to now",
			mockClient: &mockDockerClient{
				containerInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.InspectResponse, error) {
					return container.InspectResponse{
						ContainerJSONBase: &container.ContainerJSONBase{
							ID:      "bad-time-id",
							Name:    "/bad-time",
							State:   &container.State{Status: "running"},
							Created: "not-a-valid-timestamp",
						},
						Config: &container.Config{Image: "nginx:latest"},
					}, nil
				},
			},
			containerID: "bad-time-id",
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.NoError(err)
				s.NotNil(detail)
				s.WithinDuration(time.Now(), detail.Created, 5*time.Second)
			},
		},
		{
			name: "with network settings",
			mockClient: &mockDockerClient{
				containerInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.InspectResponse, error) {
					return container.InspectResponse{
						ContainerJSONBase: &container.ContainerJSONBase{
							ID:      "net-id",
							Name:    "/net-container",
							State:   &container.State{Status: "running"},
							Created: time.Now().Format(time.RFC3339Nano),
						},
						Config: &container.Config{Image: "nginx:latest"},
						NetworkSettings: &container.NetworkSettings{
							Networks: map[string]*network.EndpointSettings{
								"bridge": {
									IPAddress: "172.17.0.2",
									Gateway:   "172.17.0.1",
								},
							},
						},
					}, nil
				},
			},
			containerID: "net-id",
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.NoError(err)
				s.NotNil(detail)
				s.NotNil(detail.NetworkSettings)
				s.Equal("172.17.0.2", detail.NetworkSettings.IPAddress)
				s.Equal("172.17.0.1", detail.NetworkSettings.Gateway)
			},
		},
		{
			name: "with port bindings",
			mockClient: &mockDockerClient{
				containerInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.InspectResponse, error) {
					return container.InspectResponse{
						ContainerJSONBase: &container.ContainerJSONBase{
							ID:   "ports-id",
							Name: "/ports-container",
							State: &container.State{
								Status: "running",
							},
							Created: time.Now().Format(time.RFC3339Nano),
							HostConfig: &container.HostConfig{
								PortBindings: nat.PortMap{
									"80/tcp": []nat.PortBinding{
										{HostPort: "8080"},
									},
								},
							},
						},
						Config: &container.Config{Image: "nginx:latest"},
					}, nil
				},
			},
			containerID: "ports-id",
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.NoError(err)
				s.NotNil(detail)
				s.Len(detail.Ports, 1)
				s.Equal(8080, detail.Ports[0].Host)
				s.Equal(80, detail.Ports[0].Container)
			},
		},
		{
			name: "with mounts",
			mockClient: &mockDockerClient{
				containerInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.InspectResponse, error) {
					return container.InspectResponse{
						ContainerJSONBase: &container.ContainerJSONBase{
							ID:      "mounts-id",
							Name:    "/mounts-container",
							State:   &container.State{Status: "running"},
							Created: time.Now().Format(time.RFC3339Nano),
						},
						Config: &container.Config{Image: "nginx:latest"},
						Mounts: []container.MountPoint{
							{
								Source:      "/host/path",
								Destination: "/container/path",
							},
						},
					}, nil
				},
			},
			containerID: "mounts-id",
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.NoError(err)
				s.NotNil(detail)
				s.Len(detail.Mounts, 1)
				s.Equal("/host/path", detail.Mounts[0].Host)
				s.Equal("/container/path", detail.Mounts[0].Container)
			},
		},
		{
			name: "with health status",
			mockClient: &mockDockerClient{
				containerInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.InspectResponse, error) {
					return container.InspectResponse{
						ContainerJSONBase: &container.ContainerJSONBase{
							ID:   "health-id",
							Name: "/health-container",
							State: &container.State{
								Status: "running",
								Health: &container.Health{
									Status: "healthy",
								},
							},
							Created: time.Now().Format(time.RFC3339Nano),
						},
						Config: &container.Config{Image: "nginx:latest"},
					}, nil
				},
			},
			containerID: "health-id",
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.NoError(err)
				s.NotNil(detail)
				s.Equal("healthy", detail.Health)
			},
		},
		{
			name: "returns error when inspect fails",
			mockClient: &mockDockerClient{
				containerInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.InspectResponse, error) {
					return container.InspectResponse{}, fmt.Errorf("no such container")
				},
			},
			containerID: "missing-id",
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.Error(err)
				s.Nil(detail)
				s.Contains(err.Error(), "inspect container")
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

func (s *DockerDriverPublicTestSuite) TestExec() {
	tests := []struct {
		name         string
		mockClient   *mockDockerClient
		containerID  string
		params       runtime.ExecParams
		validateFunc func(result *runtime.ExecResult, err error)
	}{
		{
			name: "successful exec",
			mockClient: &mockDockerClient{
				containerExecCreateFunc: func(
					_ context.Context,
					_ string,
					config container.ExecOptions,
				) (container.ExecCreateResponse, error) {
					s.Equal([]string{"echo", "hello"}, config.Cmd)
					s.True(config.AttachStdout)
					s.True(config.AttachStderr)

					return container.ExecCreateResponse{ID: "exec-id"}, nil
				},
				containerExecAttachFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecStartOptions,
				) (types.HijackedResponse, error) {
					return newHijackedResponse("hello\n"), nil
				},
				containerExecInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.ExecInspect, error) {
					return container.ExecInspect{ExitCode: 0}, nil
				},
			},
			containerID: "test-id",
			params: runtime.ExecParams{
				Command: []string{"echo", "hello"},
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.NoError(err)
				s.NotNil(result)
				s.Equal("hello\n", result.Stdout)
				s.Equal(0, result.ExitCode)
			},
		},
		{
			name: "with env set",
			mockClient: &mockDockerClient{
				containerExecCreateFunc: func(
					_ context.Context,
					_ string,
					config container.ExecOptions,
				) (container.ExecCreateResponse, error) {
					s.Len(config.Env, 1)
					s.Contains(config.Env[0], "MY_VAR=value")

					return container.ExecCreateResponse{ID: "exec-env-id"}, nil
				},
				containerExecAttachFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecStartOptions,
				) (types.HijackedResponse, error) {
					return newHijackedResponse(""), nil
				},
				containerExecInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.ExecInspect, error) {
					return container.ExecInspect{ExitCode: 0}, nil
				},
			},
			containerID: "test-id",
			params: runtime.ExecParams{
				Command: []string{"env"},
				Env:     map[string]string{"MY_VAR": "value"},
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.NoError(err)
				s.NotNil(result)
			},
		},
		{
			name: "with working dir set",
			mockClient: &mockDockerClient{
				containerExecCreateFunc: func(
					_ context.Context,
					_ string,
					config container.ExecOptions,
				) (container.ExecCreateResponse, error) {
					s.Equal("/app", config.WorkingDir)

					return container.ExecCreateResponse{ID: "exec-wd-id"}, nil
				},
				containerExecAttachFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecStartOptions,
				) (types.HijackedResponse, error) {
					return newHijackedResponse(""), nil
				},
				containerExecInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.ExecInspect, error) {
					return container.ExecInspect{ExitCode: 0}, nil
				},
			},
			containerID: "test-id",
			params: runtime.ExecParams{
				Command:    []string{"pwd"},
				WorkingDir: "/app",
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.NoError(err)
				s.NotNil(result)
			},
		},
		{
			name: "returns error when exec create fails",
			mockClient: &mockDockerClient{
				containerExecCreateFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecOptions,
				) (container.ExecCreateResponse, error) {
					return container.ExecCreateResponse{}, fmt.Errorf("container not running")
				},
			},
			containerID: "stopped-id",
			params: runtime.ExecParams{
				Command: []string{"ls"},
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "create exec")
			},
		},
		{
			name: "returns error when exec attach fails",
			mockClient: &mockDockerClient{
				containerExecCreateFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecOptions,
				) (container.ExecCreateResponse, error) {
					return container.ExecCreateResponse{ID: "exec-attach-fail"}, nil
				},
				containerExecAttachFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecStartOptions,
				) (types.HijackedResponse, error) {
					return types.HijackedResponse{}, fmt.Errorf("attach failed")
				},
			},
			containerID: "test-id",
			params: runtime.ExecParams{
				Command: []string{"ls"},
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "attach exec")
			},
		},
		{
			name: "returns error when io.Copy read fails",
			mockClient: &mockDockerClient{
				containerExecCreateFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecOptions,
				) (container.ExecCreateResponse, error) {
					return container.ExecCreateResponse{ID: "exec-read-fail"}, nil
				},
				containerExecAttachFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecStartOptions,
				) (types.HijackedResponse, error) {
					return newErrorHijackedResponse(), nil
				},
			},
			containerID: "container-1",
			params: runtime.ExecParams{
				Command: []string{"ls"},
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "read exec output")
			},
		},
		{
			name: "returns error when exec inspect fails",
			mockClient: &mockDockerClient{
				containerExecCreateFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecOptions,
				) (container.ExecCreateResponse, error) {
					return container.ExecCreateResponse{ID: "exec-inspect-fail"}, nil
				},
				containerExecAttachFunc: func(
					_ context.Context,
					_ string,
					_ container.ExecStartOptions,
				) (types.HijackedResponse, error) {
					return newHijackedResponse(""), nil
				},
				containerExecInspectFunc: func(
					_ context.Context,
					_ string,
				) (container.ExecInspect, error) {
					return container.ExecInspect{}, fmt.Errorf("inspect failed")
				},
			},
			containerID: "test-id",
			params: runtime.ExecParams{
				Command: []string{"ls"},
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "inspect exec")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := docker.NewWithClient(tt.mockClient)
			result, err := d.Exec(s.ctx, tt.containerID, tt.params)
			tt.validateFunc(result, err)
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
					_ context.Context,
					_ string,
					_ image.PullOptions,
				) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("{}")), nil
				},
				imageInspectFunc: func(
					_ context.Context,
					_ string,
					_ ...dockerclient.ImageInspectOption,
				) (image.InspectResponse, error) {
					return image.InspectResponse{
						ID:       "sha256:test123",
						RepoTags: []string{"nginx:latest"},
						Size:     2048,
					}, nil
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
		{
			name: "successful pull with status digest in last event",
			mockClient: &mockDockerClient{
				imagePullFunc: func(
					_ context.Context,
					_ string,
					_ image.PullOptions,
				) (io.ReadCloser, error) {
					events := `{"status":"Pulling from library/nginx"}
{"status":"Status","id":"sha256:digest-from-event"}`

					return io.NopCloser(strings.NewReader(events)), nil
				},
				imageInspectFunc: func(
					_ context.Context,
					_ string,
					_ ...dockerclient.ImageInspectOption,
				) (image.InspectResponse, error) {
					return image.InspectResponse{
						ID:       "sha256:inspect-id",
						RepoTags: []string{"nginx:1.25"},
						Size:     4096,
					}, nil
				},
			},
			imageName: "nginx:1.25",
			validateFunc: func(
				result *runtime.PullResult,
				err error,
			) {
				s.NoError(err)
				s.NotNil(result)
				s.Equal("1.25", result.Tag)
				s.Equal(int64(4096), result.Size)
			},
		},
		{
			name: "returns error when pull fails",
			mockClient: &mockDockerClient{
				imagePullFunc: func(
					_ context.Context,
					_ string,
					_ image.PullOptions,
				) (io.ReadCloser, error) {
					return nil, fmt.Errorf("registry unavailable")
				},
			},
			imageName: "nonexistent:latest",
			validateFunc: func(
				result *runtime.PullResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "pull image")
			},
		},
		{
			name: "returns error on decode failure",
			mockClient: &mockDockerClient{
				imagePullFunc: func(
					_ context.Context,
					_ string,
					_ image.PullOptions,
				) (io.ReadCloser, error) {
					// Valid JSON followed by malformed JSON
					return io.NopCloser(strings.NewReader("{}\n{invalid json")), nil
				},
			},
			imageName: "nginx:latest",
			validateFunc: func(
				result *runtime.PullResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "decode pull response")
			},
		},
		{
			name: "image with no repo tags defaults to latest",
			mockClient: &mockDockerClient{
				imagePullFunc: func(
					_ context.Context,
					_ string,
					_ image.PullOptions,
				) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("{}")), nil
				},
				imageInspectFunc: func(
					_ context.Context,
					_ string,
					_ ...dockerclient.ImageInspectOption,
				) (image.InspectResponse, error) {
					return image.InspectResponse{
						ID:       "sha256:no-tags",
						RepoTags: []string{},
						Size:     512,
					}, nil
				},
			},
			imageName: "custom-image",
			validateFunc: func(
				result *runtime.PullResult,
				err error,
			) {
				s.NoError(err)
				s.NotNil(result)
				s.Equal("latest", result.Tag)
			},
		},
		{
			name: "returns error when image inspect fails",
			mockClient: &mockDockerClient{
				imagePullFunc: func(
					_ context.Context,
					_ string,
					_ image.PullOptions,
				) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("{}")), nil
				},
				imageInspectFunc: func(
					_ context.Context,
					_ string,
					_ ...dockerclient.ImageInspectOption,
				) (image.InspectResponse, error) {
					return image.InspectResponse{}, fmt.Errorf("image not found")
				},
			},
			imageName: "nginx:latest",
			validateFunc: func(
				result *runtime.PullResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "inspect pulled image")
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
