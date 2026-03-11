package container_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/container"
	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

type ProviderPublicTestSuite struct {
	suite.Suite
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

func TestProviderPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderPublicTestSuite))
}
