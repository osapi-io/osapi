package container

import (
	"context"
	"time"

	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

// Service implements Provider by delegating to a runtime.Driver.
type Service struct {
	driver runtime.Driver
}

// New creates a new container provider service.
func New(
	driver runtime.Driver,
) *Service {
	return &Service{driver: driver}
}

// Create delegates to the driver to create a container.
func (
	s *Service,
) Create(
	ctx context.Context,
	params runtime.CreateParams,
) (*runtime.Container, error) {
	return s.driver.Create(ctx, params)
}

// Start delegates to the driver to start a container.
func (
	s *Service,
) Start(
	ctx context.Context,
	id string,
) error {
	return s.driver.Start(ctx, id)
}

// Stop delegates to the driver to stop a container.
func (
	s *Service,
) Stop(
	ctx context.Context,
	id string,
	timeout *time.Duration,
) error {
	return s.driver.Stop(ctx, id, timeout)
}

// Remove delegates to the driver to remove a container.
func (
	s *Service,
) Remove(
	ctx context.Context,
	id string,
	force bool,
) error {
	return s.driver.Remove(ctx, id, force)
}

// List delegates to the driver to list containers.
func (
	s *Service,
) List(
	ctx context.Context,
	params runtime.ListParams,
) ([]runtime.Container, error) {
	return s.driver.List(ctx, params)
}

// Inspect delegates to the driver to inspect a container.
func (
	s *Service,
) Inspect(
	ctx context.Context,
	id string,
) (*runtime.ContainerDetail, error) {
	return s.driver.Inspect(ctx, id)
}

// Exec delegates to the driver to execute a command in a container.
func (
	s *Service,
) Exec(
	ctx context.Context,
	id string,
	params runtime.ExecParams,
) (*runtime.ExecResult, error) {
	return s.driver.Exec(ctx, id, params)
}

// Pull delegates to the driver to pull a container image.
func (
	s *Service,
) Pull(
	ctx context.Context,
	image string,
) (*runtime.PullResult, error) {
	return s.driver.Pull(ctx, image)
}
