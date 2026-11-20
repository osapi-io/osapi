// Copyright (c) 2026 John Dewey

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

package agent_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	providermocks "github.com/osapi-io/osapi/internal/provider/mocks"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	filemocks "github.com/osapi-io/osapi/internal/provider/file/mocks"
)

type ProvidersPublicTestSuite struct {
	suite.Suite
}

func TestProvidersPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProvidersPublicTestSuite))
}

func (s *ProvidersPublicTestSuite) TestProviders() {
	areas := []string{
		"host", "disk", "memory", "load", "dns", "ping", "netinfo", "cron",
		"ntp", "timezone", "power", "process", "user", "packages",
		"logs", "certificate", "service", "iface", "route",
	}

	tests := []struct {
		name      string
		platform  string
		container bool
		buckets   bool
		// want is the word that the type of each provider has. except
		// changes it for some areas.
		want   string
		except map[string]string
	}{
		{
			name:     "when the host is Debian uses the Debian providers",
			platform: "debian",
			buckets:  true,
			want:     "Debian",
			except:   map[string]string{"netinfo": "Linux"},
		},
		{
			name:      "when Debian runs in a container leaves out what a container cannot do",
			platform:  "debian",
			container: true,
			buckets:   true,
			want:      "Debian",
			except: map[string]string{
				"host":     "DebianDocker",
				"dns":      "DebianDocker",
				"netinfo":  "Linux",
				"ntp":      "Linux",
				"timezone": "Linux",
				"power":    "Linux",
				"logs":     "Linux",
				"service":  "Linux",
			},
		},
		{
			name:     "when Debian has no file buckets leaves out what needs them",
			platform: "debian",
			want:     "Debian",
			except: map[string]string{
				"netinfo":     "Linux",
				"cron":        "Linux",
				"certificate": "Linux",
				"service":     "Linux",
				"iface":       "Linux",
				"route":       "Linux",
			},
		},
		{
			name:     "when the host is macOS uses the Darwin providers",
			platform: "darwin",
			buckets:  true,
			want:     "Darwin",
		},
		{
			name:     "when the platform is another one uses the Linux fallback",
			platform: "rhel",
			buckets:  true,
			want:     "Linux",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			var providers map[string]any
			if tc.buckets {
				providers = agent.ExportProviders(
					slog.Default(),
					tc.platform,
					tc.container,
					providermocks.NewMockFileState(ctrl),
					filemocks.NewMockProvider(ctrl),
				)
			} else {
				providers = agent.ExportProviders(
					slog.Default(),
					tc.platform,
					tc.container,
					nil,
					nil,
				)
			}

			for _, area := range areas {
				want := tc.want
				if w, ok := tc.except[area]; ok {
					want = w
				}
				got := fmt.Sprintf("%T", providers[area])
				name := got[strings.LastIndex(got, ".")+1:]
				s.True(
					strings.HasPrefix(name, want) &&
						(want != "Debian" || !strings.HasPrefix(name, "DebianDocker")),
					"%s: %s, want %s", area, got, want,
				)
			}
		})
	}
}
