// Package container provides the container management provider.
package container

import (
	"context"
	"time"

	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

// Provider defines the container management interface.
// All methods accept context.Context for cancellation and timeout propagation,
// which is important since the Docker daemon is a remote service.
type Provider interface {
	Create(
		ctx context.Context,
		params runtime.CreateParams,
	) (*runtime.Container, error)

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
		params runtime.ListParams,
	) ([]runtime.Container, error)

	Inspect(
		ctx context.Context,
		id string,
	) (*runtime.ContainerDetail, error)

	Exec(
		ctx context.Context,
		id string,
		params runtime.ExecParams,
	) (*runtime.ExecResult, error)

	Pull(
		ctx context.Context,
		image string,
	) (*runtime.PullResult, error)
}
