// Package docker implements the runtime.Driver interface using the Docker Engine API.
package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

// Driver implements runtime.Driver using the Docker Engine API.
type Driver struct {
	client dockerclient.APIClient
}

// New creates a new Docker driver using default client options.
func New() (*Driver, error) {
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &Driver{client: cli}, nil
}

// NewWithClient creates a Docker driver with an injected client (for testing).
func NewWithClient(
	client dockerclient.APIClient,
) *Driver {
	return &Driver{client: client}
}

// Ping verifies connectivity to the Docker daemon.
func (d *Driver) Ping(
	ctx context.Context,
) error {
	_, err := d.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping docker daemon: %w", err)
	}

	return nil
}

// Create creates a new container from the given parameters.
func (d *Driver) Create(
	ctx context.Context,
	params runtime.CreateParams,
) (*runtime.Container, error) {
	// Build Docker container configuration
	config := &container.Config{
		Image: params.Image,
	}

	if len(params.Command) > 0 {
		config.Cmd = params.Command
	}

	if len(params.Env) > 0 {
		envVars := make([]string, 0, len(params.Env))
		for k, v := range params.Env {
			envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
		}
		config.Env = envVars
	}

	// Build host configuration
	hostConfig := &container.HostConfig{}

	// Convert port mappings
	if len(params.Ports) > 0 {
		portBindings := nat.PortMap{}
		exposedPorts := nat.PortSet{}

		for _, pm := range params.Ports {
			containerPort := nat.Port(fmt.Sprintf("%d/tcp", pm.Container))
			exposedPorts[containerPort] = struct{}{}
			portBindings[containerPort] = []nat.PortBinding{
				{
					HostPort: strconv.Itoa(pm.Host),
				},
			}
		}

		config.ExposedPorts = exposedPorts
		hostConfig.PortBindings = portBindings
	}

	// Convert volume mappings
	if len(params.Volumes) > 0 {
		mounts := make([]mount.Mount, 0, len(params.Volumes))
		for _, vm := range params.Volumes {
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: vm.Host,
				Target: vm.Container,
			})
		}
		hostConfig.Mounts = mounts
	}

	// Create the container
	resp, err := d.client.ContainerCreate(
		ctx,
		config,
		hostConfig,
		&network.NetworkingConfig{},
		nil,
		params.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	// Start the container if AutoStart is enabled
	if params.AutoStart {
		if err := d.Start(ctx, resp.ID); err != nil {
			return nil, fmt.Errorf("auto-start container: %w", err)
		}
	}

	// Return container summary
	return &runtime.Container{
		ID:      resp.ID,
		Name:    params.Name,
		Image:   params.Image,
		State:   "created",
		Created: time.Now(),
	}, nil
}

// Start starts a stopped container.
func (d *Driver) Start(
	ctx context.Context,
	id string,
) error {
	if err := d.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	return nil
}

// Stop stops a running container with an optional timeout.
func (d *Driver) Stop(
	ctx context.Context,
	id string,
	timeout *time.Duration,
) error {
	opts := container.StopOptions{}
	if timeout != nil {
		seconds := int(timeout.Seconds())
		opts.Timeout = &seconds
	}

	if err := d.client.ContainerStop(ctx, id, opts); err != nil {
		return fmt.Errorf("stop container: %w", err)
	}

	return nil
}

// Remove removes a container.
func (d *Driver) Remove(
	ctx context.Context,
	id string,
	force bool,
) error {
	opts := container.RemoveOptions{
		Force: force,
	}

	if err := d.client.ContainerRemove(ctx, id, opts); err != nil {
		return fmt.Errorf("remove container: %w", err)
	}

	return nil
}

// List returns a list of containers matching the given parameters.
func (d *Driver) List(
	ctx context.Context,
	params runtime.ListParams,
) ([]runtime.Container, error) {
	opts := container.ListOptions{}

	// Apply state filter
	switch params.State {
	case "running":
		opts.All = false
	case "stopped":
		opts.All = true
		opts.Filters = filters.NewArgs(filters.Arg("status", "exited"))
	case "all":
		opts.All = true
	default:
		opts.All = false
	}

	// Apply limit
	if params.Limit > 0 {
		opts.Limit = params.Limit
	}

	containers, err := d.client.ContainerList(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	// Convert to runtime.Container
	result := make([]runtime.Container, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			// Remove leading "/" from container name
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		result = append(result, runtime.Container{
			ID:      c.ID,
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Created: time.Unix(c.Created, 0),
		})
	}

	return result, nil
}

// Inspect returns detailed information about a container.
func (d *Driver) Inspect(
	ctx context.Context,
	id string,
) (*runtime.ContainerDetail, error) {
	resp, err := d.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("inspect container: %w", err)
	}

	// Build container detail
	name := strings.TrimPrefix(resp.Name, "/")
	state := "unknown"
	if resp.State != nil {
		state = resp.State.Status
	}

	// Parse Created timestamp
	created, err := time.Parse(time.RFC3339Nano, resp.Created)
	if err != nil {
		created = time.Now()
	}

	detail := &runtime.ContainerDetail{
		Container: runtime.Container{
			ID:      resp.ID,
			Name:    name,
			Image:   resp.Config.Image,
			State:   state,
			Created: created,
		},
	}

	// Add network settings
	if resp.NetworkSettings != nil {
		for _, netConfig := range resp.NetworkSettings.Networks {
			detail.NetworkSettings = &runtime.NetworkSettings{
				IPAddress: netConfig.IPAddress,
				Gateway:   netConfig.Gateway,
			}
			break // Use first network
		}
	}

	// Add port mappings
	if resp.HostConfig != nil && resp.HostConfig.PortBindings != nil {
		ports := make([]runtime.PortMapping, 0)
		for containerPort, bindings := range resp.HostConfig.PortBindings {
			for _, binding := range bindings {
				hostPort, _ := strconv.Atoi(binding.HostPort)
				cPort, _ := strconv.Atoi(containerPort.Port())
				ports = append(ports, runtime.PortMapping{
					Host:      hostPort,
					Container: cPort,
				})
			}
		}
		detail.Ports = ports
	}

	// Add mounts
	if len(resp.Mounts) > 0 {
		mounts := make([]runtime.VolumeMapping, 0, len(resp.Mounts))
		for _, m := range resp.Mounts {
			mounts = append(mounts, runtime.VolumeMapping{
				Host:      m.Source,
				Container: m.Destination,
			})
		}
		detail.Mounts = mounts
	}

	// Add health status
	if resp.State != nil && resp.State.Health != nil {
		detail.Health = resp.State.Health.Status
	}

	return detail, nil
}

// Exec executes a command in a running container.
func (d *Driver) Exec(
	ctx context.Context,
	id string,
	params runtime.ExecParams,
) (*runtime.ExecResult, error) {
	// Create exec configuration
	execConfig := container.ExecOptions{
		Cmd:          params.Command,
		AttachStdout: true,
		AttachStderr: true,
	}

	if len(params.Env) > 0 {
		envVars := make([]string, 0, len(params.Env))
		for k, v := range params.Env {
			envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
		}
		execConfig.Env = envVars
	}

	if params.WorkingDir != "" {
		execConfig.WorkingDir = params.WorkingDir
	}

	// Create exec instance
	execResp, err := d.client.ContainerExecCreate(ctx, id, execConfig)
	if err != nil {
		return nil, fmt.Errorf("create exec: %w", err)
	}

	// Attach to exec instance
	attachResp, err := d.client.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("attach exec: %w", err)
	}
	defer attachResp.Close()

	// Read output
	var stdout, stderr bytes.Buffer
	if _, err := io.Copy(&stdout, attachResp.Reader); err != nil {
		return nil, fmt.Errorf("read exec output: %w", err)
	}

	// Get exit code
	inspectResp, err := d.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("inspect exec: %w", err)
	}

	return &runtime.ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: inspectResp.ExitCode,
	}, nil
}

// Pull pulls a container image from a registry.
func (d *Driver) Pull(
	ctx context.Context,
	imageName string,
) (*runtime.PullResult, error) {
	pullResp, err := d.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("pull image: %w", err)
	}
	defer pullResp.Close()

	// Read pull output to ensure completion
	var lastEvent map[string]interface{}
	decoder := json.NewDecoder(pullResp)
	for {
		var event map[string]interface{}
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("decode pull response: %w", err)
		}
		lastEvent = event
	}

	// Inspect the pulled image to get details
	inspectResp, _, err := d.client.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return nil, fmt.Errorf("inspect pulled image: %w", err)
	}

	tag := "latest"
	if len(inspectResp.RepoTags) > 0 {
		parts := strings.Split(inspectResp.RepoTags[0], ":")
		if len(parts) > 1 {
			tag = parts[1]
		}
	}

	result := &runtime.PullResult{
		ImageID: inspectResp.ID,
		Tag:     tag,
		Size:    inspectResp.Size,
	}

	// Extract digest from last event if available
	if lastEvent != nil {
		if status, ok := lastEvent["status"].(string); ok && status == "Status" {
			if digest, ok := lastEvent["id"].(string); ok {
				result.ImageID = digest
			}
		}
	}

	return result, nil
}
