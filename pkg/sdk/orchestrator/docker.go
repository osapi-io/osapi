package orchestrator

import (
	"context"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// DockerPull creates a task that pulls a Docker image on the target host.
func (p *Plan) DockerPull(
	name string,
	target string,
	image string,
) *Task {
	return p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		resp, err := c.Docker.Pull(ctx, target, gen.DockerPullRequest{
			Image: image,
		})
		if err != nil {
			return nil, err
		}

		r := resp.Data.Results[0]

		return &Result{
			JobID:   resp.Data.JobID,
			Changed: r.Changed,
			Data: map[string]any{
				"image_id": r.ImageID,
				"tag":      r.Tag,
				"size":     r.Size,
			},
		}, nil
	})
}

// DockerCreate creates a task that creates a Docker container on the
// target host.
func (p *Plan) DockerCreate(
	name string,
	target string,
	body gen.DockerCreateRequest,
) *Task {
	return p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		resp, err := c.Docker.Create(ctx, target, body)
		if err != nil {
			return nil, err
		}

		r := resp.Data.Results[0]

		return &Result{
			JobID:   resp.Data.JobID,
			Changed: r.Changed,
			Data: map[string]any{
				"id":    r.ID,
				"name":  r.Name,
				"image": r.Image,
				"state": r.State,
			},
		}, nil
	})
}

// DockerStart creates a task that starts a Docker container on the
// target host.
func (p *Plan) DockerStart(
	name string,
	target string,
	id string,
) *Task {
	return p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		resp, err := c.Docker.Start(ctx, target, id)
		if err != nil {
			return nil, err
		}

		r := resp.Data.Results[0]

		return &Result{
			JobID:   resp.Data.JobID,
			Changed: r.Changed,
			Data: map[string]any{
				"id":      r.ID,
				"message": r.Message,
			},
		}, nil
	})
}

// DockerStop creates a task that stops a Docker container on the
// target host.
func (p *Plan) DockerStop(
	name string,
	target string,
	id string,
	body gen.DockerStopRequest,
) *Task {
	return p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		resp, err := c.Docker.Stop(ctx, target, id, body)
		if err != nil {
			return nil, err
		}

		r := resp.Data.Results[0]

		return &Result{
			JobID:   resp.Data.JobID,
			Changed: r.Changed,
			Data: map[string]any{
				"id":      r.ID,
				"message": r.Message,
			},
		}, nil
	})
}

// DockerRemove creates a task that removes a Docker container from the
// target host.
func (p *Plan) DockerRemove(
	name string,
	target string,
	id string,
	params *gen.DeleteNodeContainerDockerByIDParams,
) *Task {
	return p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		resp, err := c.Docker.Remove(ctx, target, id, params)
		if err != nil {
			return nil, err
		}

		r := resp.Data.Results[0]

		return &Result{
			JobID:   resp.Data.JobID,
			Changed: r.Changed,
			Data: map[string]any{
				"id":      r.ID,
				"message": r.Message,
			},
		}, nil
	})
}

// DockerExec creates a task that executes a command in a Docker
// container.
func (p *Plan) DockerExec(
	name string,
	target string,
	id string,
	body gen.DockerExecRequest,
) *Task {
	return p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		resp, err := c.Docker.Exec(ctx, target, id, body)
		if err != nil {
			return nil, err
		}

		r := resp.Data.Results[0]

		return &Result{
			JobID:   resp.Data.JobID,
			Changed: r.Changed,
			Data: map[string]any{
				"stdout":    r.Stdout,
				"stderr":    r.Stderr,
				"exit_code": r.ExitCode,
			},
		}, nil
	})
}

// DockerInspect creates a task that inspects a Docker container on the
// target host.
func (p *Plan) DockerInspect(
	name string,
	target string,
	id string,
) *Task {
	return p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		resp, err := c.Docker.Inspect(ctx, target, id)
		if err != nil {
			return nil, err
		}

		r := resp.Data.Results[0]

		return &Result{
			JobID:   resp.Data.JobID,
			Changed: false,
			Data: map[string]any{
				"id":    r.ID,
				"name":  r.Name,
				"image": r.Image,
				"state": r.State,
			},
		}, nil
	})
}

// DockerList creates a task that lists Docker containers on the target
// host.
func (p *Plan) DockerList(
	name string,
	target string,
	params *gen.GetNodeContainerDockerParams,
) *Task {
	return p.TaskFunc(name, func(
		ctx context.Context,
		c *osapiclient.Client,
	) (*Result, error) {
		resp, err := c.Docker.List(ctx, target, params)
		if err != nil {
			return nil, err
		}

		r := resp.Data.Results[0]

		return &Result{
			JobID:   resp.Data.JobID,
			Changed: false,
			Data: map[string]any{
				"containers": r.Containers,
			},
		}, nil
	})
}
