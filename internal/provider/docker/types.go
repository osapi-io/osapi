// Package docker provides the Docker container management provider.
package docker

import (
	"context"
	"time"
)

// Provider defines the Docker container management interface.
// All methods accept context.Context for cancellation and timeout propagation,
// which is important since the Docker daemon is a remote service.
type Provider interface {
	Ping(
		ctx context.Context,
	) error

	Create(
		ctx context.Context,
		params CreateParams,
	) (*Container, error)

	Start(
		ctx context.Context,
		id string,
	) error

	Stop(
		ctx context.Context,
		id string,
		timeout *time.Duration,
	) error

	Remove(
		ctx context.Context,
		id string,
		force bool,
	) error

	List(
		ctx context.Context,
		params ListParams,
	) ([]Container, error)

	Inspect(
		ctx context.Context,
		id string,
	) (*ContainerDetail, error)

	Exec(
		ctx context.Context,
		id string,
		params ExecParams,
	) (*ExecResult, error)

	Pull(
		ctx context.Context,
		image string,
	) (*PullResult, error)
}

// CreateParams contains parameters for container creation.
type CreateParams struct {
	// Image is the container image (required).
	Image string `json:"image"`
	// Name is an optional container name.
	Name string `json:"name,omitempty"`
	// Command overrides the image's default command.
	Command []string `json:"command,omitempty"`
	// Env sets environment variables.
	Env map[string]string `json:"env,omitempty"`
	// Ports maps host ports to container ports.
	Ports []PortMapping `json:"ports,omitempty"`
	// Volumes maps host paths to container paths.
	Volumes []VolumeMapping `json:"volumes,omitempty"`
	// AutoStart starts the container after creation.
	AutoStart bool `json:"auto_start,omitempty"`
}

// PortMapping maps a host port to a container port.
type PortMapping struct {
	Host      int `json:"host"`
	Container int `json:"container"`
}

// VolumeMapping maps a host path to a container path.
type VolumeMapping struct {
	Host      string `json:"host"`
	Container string `json:"container"`
}

// ListParams contains parameters for listing containers.
type ListParams struct {
	// State filters by container state: "running", "stopped", "all".
	State string `json:"state,omitempty"`
	// Limit caps the number of results.
	Limit int `json:"limit,omitempty"`
}

// Container holds summary info for a container.
type Container struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Image   string    `json:"image"`
	State   string    `json:"state"`
	Created time.Time `json:"created"`
}

// ContainerDetail holds detailed info for a container.
type ContainerDetail struct {
	Container
	NetworkSettings *NetworkSettings `json:"network_settings,omitempty"`
	Ports           []PortMapping    `json:"ports,omitempty"`
	Mounts          []VolumeMapping  `json:"mounts,omitempty"`
	Health          string           `json:"health,omitempty"`
}

// NetworkSettings holds container network configuration.
type NetworkSettings struct {
	IPAddress string `json:"ip_address,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
}

// ExecParams contains parameters for executing a command in a container.
type ExecParams struct {
	// Command is the command and arguments.
	Command []string `json:"command"`
	// Env sets environment variables.
	Env map[string]string `json:"env,omitempty"`
	// WorkingDir sets the working directory.
	WorkingDir string `json:"working_dir,omitempty"`
}

// ExecResult contains the output of a command execution in a container.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// PullResult contains the result of an image pull.
type PullResult struct {
	ImageID string `json:"image_id"`
	Tag     string `json:"tag"`
	Size    int64  `json:"size"`
}
