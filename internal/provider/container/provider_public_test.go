package container_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/container"
	"github.com/retr0h/osapi/internal/provider/container/runtime"
	runtimeMocks "github.com/retr0h/osapi/internal/provider/container/runtime/mocks"
)

type ProviderPublicTestSuite struct {
	suite.Suite

	mockCtrl   *gomock.Controller
	mockDriver *runtimeMocks.MockDriver
	service    *container.Service
	ctx        context.Context
}

func (s *ProviderPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockDriver = runtimeMocks.NewMockDriver(s.mockCtrl)
	s.service = container.New(s.mockDriver)
	s.ctx = context.Background()
}

func (s *ProviderPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (
	s *ProviderPublicTestSuite,
) TestNew() {
	tests := []struct {
		name         string
		validateFunc func(p container.Provider)
	}{
		{
			name: "returns non-nil provider",
			validateFunc: func(p container.Provider) {
				s.NotNil(p)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var driver runtime.Driver // nil driver for unit test
			p := container.New(driver)
			tt.validateFunc(p)
		})
	}
}

func (
	s *ProviderPublicTestSuite,
) TestCreate() {
	tests := []struct {
		name         string
		params       runtime.CreateParams
		setupMock    func()
		validateFunc func(*runtime.Container, error)
	}{
		{
			name: "delegates to driver and returns result",
			params: runtime.CreateParams{
				Image: "nginx:latest",
				Name:  "web",
			},
			setupMock: func() {
				s.mockDriver.EXPECT().
					Create(gomock.Any(), runtime.CreateParams{
						Image: "nginx:latest",
						Name:  "web",
					}).
					Return(&runtime.Container{
						ID:    "abc123",
						Name:  "web",
						Image: "nginx:latest",
						State: "created",
					}, nil)
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.NotNil(c)
				s.Equal("abc123", c.ID)
				s.Equal("web", c.Name)
			},
		},
		{
			name: "returns error from driver",
			params: runtime.CreateParams{
				Image: "invalid:image",
			},
			setupMock: func() {
				s.mockDriver.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("image not found"))
			},
			validateFunc: func(
				c *runtime.Container,
				err error,
			) {
				s.Error(err)
				s.Nil(c)
				s.Contains(err.Error(), "image not found")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			c, err := s.service.Create(s.ctx, tt.params)
			tt.validateFunc(c, err)
		})
	}
}

func (
	s *ProviderPublicTestSuite,
) TestStart() {
	tests := []struct {
		name         string
		id           string
		setupMock    func()
		validateFunc func(error)
	}{
		{
			name: "delegates to driver and returns nil",
			id:   "abc123",
			setupMock: func() {
				s.mockDriver.EXPECT().
					Start(gomock.Any(), "abc123").
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "returns error from driver",
			id:   "abc123",
			setupMock: func() {
				s.mockDriver.EXPECT().
					Start(gomock.Any(), "abc123").
					Return(errors.New("container not found"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "container not found")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			err := s.service.Start(s.ctx, tt.id)
			tt.validateFunc(err)
		})
	}
}

func (
	s *ProviderPublicTestSuite,
) TestStop() {
	timeout := 10 * time.Second
	tests := []struct {
		name         string
		id           string
		timeout      *time.Duration
		setupMock    func()
		validateFunc func(error)
	}{
		{
			name:    "delegates to driver with timeout",
			id:      "abc123",
			timeout: &timeout,
			setupMock: func() {
				s.mockDriver.EXPECT().
					Stop(gomock.Any(), "abc123", &timeout).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:    "delegates to driver without timeout",
			id:      "abc123",
			timeout: nil,
			setupMock: func() {
				s.mockDriver.EXPECT().
					Stop(gomock.Any(), "abc123", (*time.Duration)(nil)).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:    "returns error from driver",
			id:      "abc123",
			timeout: &timeout,
			setupMock: func() {
				s.mockDriver.EXPECT().
					Stop(gomock.Any(), "abc123", gomock.Any()).
					Return(errors.New("stop failed"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "stop failed")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			err := s.service.Stop(s.ctx, tt.id, tt.timeout)
			tt.validateFunc(err)
		})
	}
}

func (
	s *ProviderPublicTestSuite,
) TestRemove() {
	tests := []struct {
		name         string
		id           string
		force        bool
		setupMock    func()
		validateFunc func(error)
	}{
		{
			name:  "delegates to driver with force",
			id:    "abc123",
			force: true,
			setupMock: func() {
				s.mockDriver.EXPECT().
					Remove(gomock.Any(), "abc123", true).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:  "delegates to driver without force",
			id:    "abc123",
			force: false,
			setupMock: func() {
				s.mockDriver.EXPECT().
					Remove(gomock.Any(), "abc123", false).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:  "returns error from driver",
			id:    "abc123",
			force: true,
			setupMock: func() {
				s.mockDriver.EXPECT().
					Remove(gomock.Any(), "abc123", true).
					Return(errors.New("remove failed"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "remove failed")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			err := s.service.Remove(s.ctx, tt.id, tt.force)
			tt.validateFunc(err)
		})
	}
}

func (
	s *ProviderPublicTestSuite,
) TestList() {
	tests := []struct {
		name         string
		params       runtime.ListParams
		setupMock    func()
		validateFunc func([]runtime.Container, error)
	}{
		{
			name: "delegates to driver and returns containers",
			params: runtime.ListParams{
				State: "running",
				Limit: 10,
			},
			setupMock: func() {
				s.mockDriver.EXPECT().
					List(gomock.Any(), runtime.ListParams{
						State: "running",
						Limit: 10,
					}).
					Return([]runtime.Container{
						{ID: "abc123", Name: "web", State: "running"},
					}, nil)
			},
			validateFunc: func(
				containers []runtime.Container,
				err error,
			) {
				s.NoError(err)
				s.Len(containers, 1)
				s.Equal("abc123", containers[0].ID)
			},
		},
		{
			name:   "returns error from driver",
			params: runtime.ListParams{State: "all"},
			setupMock: func() {
				s.mockDriver.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("list failed"))
			},
			validateFunc: func(
				containers []runtime.Container,
				err error,
			) {
				s.Error(err)
				s.Nil(containers)
				s.Contains(err.Error(), "list failed")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			containers, err := s.service.List(s.ctx, tt.params)
			tt.validateFunc(containers, err)
		})
	}
}

func (
	s *ProviderPublicTestSuite,
) TestInspect() {
	tests := []struct {
		name         string
		id           string
		setupMock    func()
		validateFunc func(*runtime.ContainerDetail, error)
	}{
		{
			name: "delegates to driver and returns detail",
			id:   "abc123",
			setupMock: func() {
				s.mockDriver.EXPECT().
					Inspect(gomock.Any(), "abc123").
					Return(&runtime.ContainerDetail{
						Container: runtime.Container{ID: "abc123"},
					}, nil)
			},
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.NoError(err)
				s.NotNil(detail)
				s.Equal("abc123", detail.ID)
			},
		},
		{
			name: "returns error from driver",
			id:   "abc123",
			setupMock: func() {
				s.mockDriver.EXPECT().
					Inspect(gomock.Any(), "abc123").
					Return(nil, errors.New("inspect failed"))
			},
			validateFunc: func(
				detail *runtime.ContainerDetail,
				err error,
			) {
				s.Error(err)
				s.Nil(detail)
				s.Contains(err.Error(), "inspect failed")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			detail, err := s.service.Inspect(s.ctx, tt.id)
			tt.validateFunc(detail, err)
		})
	}
}

func (
	s *ProviderPublicTestSuite,
) TestExec() {
	tests := []struct {
		name         string
		id           string
		params       runtime.ExecParams
		setupMock    func()
		validateFunc func(*runtime.ExecResult, error)
	}{
		{
			name: "delegates to driver and returns result",
			id:   "abc123",
			params: runtime.ExecParams{
				Command: []string{"ls", "-la"},
			},
			setupMock: func() {
				s.mockDriver.EXPECT().
					Exec(gomock.Any(), "abc123", runtime.ExecParams{
						Command: []string{"ls", "-la"},
					}).
					Return(&runtime.ExecResult{
						Stdout:   "output",
						ExitCode: 0,
					}, nil)
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.NoError(err)
				s.NotNil(result)
				s.Equal("output", result.Stdout)
				s.Equal(0, result.ExitCode)
			},
		},
		{
			name: "returns error from driver",
			id:   "abc123",
			params: runtime.ExecParams{
				Command: []string{"ls"},
			},
			setupMock: func() {
				s.mockDriver.EXPECT().
					Exec(gomock.Any(), "abc123", gomock.Any()).
					Return(nil, errors.New("exec failed"))
			},
			validateFunc: func(
				result *runtime.ExecResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "exec failed")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			result, err := s.service.Exec(s.ctx, tt.id, tt.params)
			tt.validateFunc(result, err)
		})
	}
}

func (
	s *ProviderPublicTestSuite,
) TestPull() {
	tests := []struct {
		name         string
		image        string
		setupMock    func()
		validateFunc func(*runtime.PullResult, error)
	}{
		{
			name:  "delegates to driver and returns result",
			image: "nginx:latest",
			setupMock: func() {
				s.mockDriver.EXPECT().
					Pull(gomock.Any(), "nginx:latest").
					Return(&runtime.PullResult{
						ImageID: "sha256:abc",
						Tag:     "latest",
						Size:    2048,
					}, nil)
			},
			validateFunc: func(
				result *runtime.PullResult,
				err error,
			) {
				s.NoError(err)
				s.NotNil(result)
				s.Equal("sha256:abc", result.ImageID)
				s.Equal("latest", result.Tag)
				s.Equal(int64(2048), result.Size)
			},
		},
		{
			name:  "returns error from driver",
			image: "invalid:image",
			setupMock: func() {
				s.mockDriver.EXPECT().
					Pull(gomock.Any(), "invalid:image").
					Return(nil, errors.New("pull failed"))
			},
			validateFunc: func(
				result *runtime.PullResult,
				err error,
			) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "pull failed")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			result, err := s.service.Pull(s.ctx, tt.image)
			tt.validateFunc(result, err)
		})
	}
}

func TestProviderPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderPublicTestSuite))
}
