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
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
)

type RotateKeyPublicTestSuite struct {
	suite.Suite
}

func TestRotateKeyPublicTestSuite(t *testing.T) {
	suite.Run(t, new(RotateKeyPublicTestSuite))
}

// failOn is fs, failing fn on a path that ends with suffix, or that holds
// it when anywhere is true.
func failOn(
	fs avfs.VFS,
	fn avfs.FnVFS,
	suffix string,
	anywhere ...bool,
) avfs.VFS {
	vfs := failfs.New(fs)
	_ = vfs.SetFailFunc(func(_ avfs.VFSBase, f avfs.FnVFS, fp *failfs.FailParam) error {
		match := strings.HasSuffix(fp.Path, suffix)
		if len(anywhere) > 0 && anywhere[0] {
			match = strings.Contains(fp.Path, suffix)
		}
		if f == fn && match {
			return errors.New("permission denied")
		}

		return nil
	})

	return vfs
}

// identityOf is the identity of a host with machine ID m-1.
func identityOf(avfs.VFS, string) (*identity.Identity, error) {
	return &identity.Identity{MachineID: "m-1", Hostname: "web-01"}, nil
}

func (s *RotateKeyPublicTestSuite) TestRotateKey() {
	// keys makes the agent's keys in /pki, with a saved join over tls when
	// joined is true, and returns the file system and the keys. The join
	// trusts the system's roots for a name, or a pinned CA when ca is true.
	keys := func(joined bool, ca bool) (avfs.VFS, *pki.Manager) {
		fs := memfs.New()
		s.Require().NoError(fs.MkdirAll("/pki", 0o700))
		m := pki.New(fs, "/pki", "agent", pki.WithEncryptionKey())
		s.Require().NoError(m.LoadOrGenerate())
		if joined {
			join := map[string]any{
				"scheme": "tls", "host": "hub", "port": 4222, "namespace": "osapi",
				"creds_file": "/pki/nats.creds", "server_name": "hub.example.com",
			}
			if ca {
				join["ca"] = "PEM"
			}
			saved, _ := json.Marshal(join)
			s.Require().NoError(avfs.WriteFile(fs, "/pki/"+agent.NATSJoinFile, saved, 0o600))
		}

		return fs, m
	}
	static := config.Config{Agent: config.AgentConfig{
		Hostname: "web-01",
		PKI:      config.AgentPKI{KeyDir: "/pki"},
		NATS: config.NATSConnection{
			Host: "nats.example.com", Port: 4222,
			Auth: config.NATSAuth{Type: "user_pass", Username: "agent", Password: "pw"},
		},
	}}
	// taken is a controller that takes the rotation, after it checks it
	// against old, the keys that it admitted.
	taken := func(old *pki.Manager, wantURL, wantSubject string) func(
		string, []nats.Option, string, []byte,
	) ([]byte, error) {
		return func(url string, _ []nats.Option, subject string, data []byte) ([]byte, error) {
			s.Equal(wantURL, url)
			s.Equal(wantSubject, subject)
			var req pki.RotateRequest
			s.Require().NoError(json.Unmarshal(data, &req))
			s.Equal("m-1", req.MachineID)
			s.True(req.Verify(old.PublicKey()), "signed with the old key and the new one")

			return json.Marshal(pki.RotateReply{
				Fingerprint: pki.FingerprintOfKeys(req.PublicKey, req.EncryptionKey),
			})
		}
	}
	refused := func(string, []nats.Option, string, []byte) ([]byte, error) {
		return json.Marshal(pki.RotateReply{Error: "the machine is not admitted"})
	}

	tests := []struct {
		name    string
		joined  bool
		ca      bool
		noKeys  bool
		fs      func(avfs.VFS) avfs.VFS
		cfg     func(config.Config) config.Config
		id      func(avfs.VFS, string) (*identity.Identity, error)
		request func(old *pki.Manager) func(string, []nats.Option, string, []byte) ([]byte, error)
		wantErr string
		// wantNext is whether the new keys stay in agent.next.* after, and
		// wantRotated whether they are the agent's keys.
		wantNext    bool
		wantRotated bool
		isRefused   bool
	}{
		{
			name:   "an agent that joined over tls rotates, as it connects",
			joined: true,
			request: func(old *pki.Manager) func(string, []nats.Option, string, []byte) ([]byte, error) {
				return taken(old, "tls://hub:4222", "osapi.join.rotate")
			},
			wantRotated: true,
		},
		{
			name: "an agent with the NATS user of its config rotates, with no namespace",
			request: func(old *pki.Manager) func(string, []nats.Option, string, []byte) ([]byte, error) {
				return taken(old, "nats://nats.example.com:4222", "join.rotate")
			},
			wantRotated: true,
		},
		{
			name:    "an agent with no keys",
			noKeys:  true,
			wantErr: "the agent makes them at its first start",
		},
		{
			name:    "the new keys cannot be made",
			fs:      func(fs avfs.VFS) avfs.VFS { return failOn(fs, avfs.FnOpenFile, ".next.key") },
			wantErr: "make the new keys",
		},
		{
			name: "the identity of the machine cannot be read",
			id: func(avfs.VFS, string) (*identity.Identity, error) {
				return nil, errors.New("no machine-id")
			},
			wantErr:  "the identity of this machine: no machine-id",
			wantNext: true,
		},
		{
			name:     "the certificate of the saved join cannot be written",
			joined:   true,
			ca:       true,
			fs:       func(fs avfs.VFS) avfs.VFS { return failOn(fs, avfs.FnOpenFile, "nats-join-ca.pem", true) },
			wantErr:  "permission denied",
			wantNext: true,
		},
		{
			name: "the TLS of the config does not load",
			cfg: func(c config.Config) config.Config {
				c.Agent.NATS.TLS = config.ClientTLS{Enabled: true, CAFile: "/nonexistent/ca.pem"}

				return c
			},
			wantErr:  "nats tls",
			wantNext: true,
		},
		{
			name: "the controller does not answer: the new keys stay for a second rotation",
			request: func(*pki.Manager) func(string, []nats.Option, string, []byte) ([]byte, error) {
				return func(string, []nats.Option, string, []byte) ([]byte, error) {
					return nil, nats.ErrTimeout
				}
			},
			wantErr:  "a second rotation sends them again",
			wantNext: true,
		},
		{
			name: "a reply that is not JSON",
			request: func(*pki.Manager) func(string, []nats.Option, string, []byte) ([]byte, error) {
				return func(string, []nats.Option, string, []byte) ([]byte, error) {
					return []byte("{"), nil
				}
			},
			wantErr:  "the reply of the rotation",
			wantNext: true,
		},
		{
			name: "the controller refuses: the new keys are removed",
			request: func(*pki.Manager) func(string, []nats.Option, string, []byte) ([]byte, error) {
				return refused
			},
			wantErr:   "the controller refused the rotation: the machine is not admitted",
			isRefused: true,
		},
		{
			name: "the controller refuses, and the new keys cannot be removed",
			fs:   func(fs avfs.VFS) avfs.VFS { return failOn(fs, avfs.FnRemove, ".next.key") },
			request: func(*pki.Manager) func(string, []nats.Option, string, []byte) ([]byte, error) {
				return refused
			},
			wantErr:   "the controller refused the rotation",
			isRefused: true,
			wantNext:  true,
		},
		{
			name: "the controller takes the keys, and they cannot be moved in place",
			fs:   func(fs avfs.VFS) avfs.VFS { return failOn(fs, avfs.FnRename, ".next.ekey") },
			request: func(old *pki.Manager) func(string, []nats.Option, string, []byte) ([]byte, error) {
				return taken(old, "nats://nats.example.com:4222", "join.rotate")
			},
			wantErr:  "run the rotation again to move them in place",
			wantNext: true,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			base, old := keys(tt.joined, tt.ca)
			if tt.noKeys {
				base = memfs.New()
			}
			fs := base
			if tt.fs != nil {
				fs = tt.fs(base)
			}
			cfg := static
			if tt.cfg != nil {
				cfg = tt.cfg(cfg)
			}
			id := tt.id
			if id == nil {
				id = identityOf
			}
			var req func(string, []nats.Option, string, []byte) ([]byte, error)
			if tt.request != nil {
				req = tt.request(old)
			} else {
				req = func(string, []nats.Option, string, []byte) ([]byte, error) {
					s.Fail("no request is sent")

					return nil, nil
				}
			}
			defer agent.SetJoinNATSFns(id, req)()

			fp, err := agent.RotateKey(context.Background(), slog.Default(), fs, cfg)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				s.Equal(tt.isRefused, errors.Is(err, agent.ErrRotateRefused))
			} else {
				s.Require().NoError(err)
			}
			if tt.noKeys {
				return
			}
			after := pki.New(base, "/pki", "agent", pki.WithEncryptionKey())
			s.Require().NoError(after.Load())
			s.Equal(tt.wantRotated, after.Fingerprint() != old.Fingerprint())
			if tt.wantRotated {
				s.Equal(after.Fingerprint(), fp)
			}
			_, statErr := base.Stat("/pki/agent.next.key")
			s.Equal(tt.wantNext, statErr == nil)
		})
	}
}

func (s *RotateKeyPublicTestSuite) TestNATSDialOptions() {
	dir := s.T().TempDir()
	user, _ := nkeys.CreateUser()
	seed, _ := user.Seed()
	seedFile := filepath.Join(dir, "agent.nk")
	s.Require().NoError(os.WriteFile(seedFile, seed, 0o600))

	tests := []struct {
		name     string
		conn     config.NATSConnection
		wantURL  string
		wantOpts int
		wantErr  string
	}{
		{
			name:     "no auth",
			conn:     config.NATSConnection{Host: "::1", Port: 4222},
			wantURL:  "nats://[::1]:4222",
			wantOpts: 2,
		},
		{
			name: "a user and password, with TLS",
			conn: config.NATSConnection{
				Host: "hub", Port: 4222, TLS: config.ClientTLS{Enabled: true},
				Auth: config.NATSAuth{Type: "user_pass", Username: "agent", Password: "pw"},
			},
			wantURL:  "tls://hub:4222",
			wantOpts: 4,
		},
		{
			name: "an nkey",
			conn: config.NATSConnection{
				Host: "hub", Port: 4222, Auth: config.NATSAuth{Type: "nkey", NKeyFile: seedFile},
			},
			wantURL:  "nats://hub:4222",
			wantOpts: 3,
		},
		{
			name: "an nkey that cannot be read",
			conn: config.NATSConnection{
				Host: "hub", Port: 4222,
				Auth: config.NATSAuth{Type: "nkey", NKeyFile: filepath.Join(dir, "missing.nk")},
			},
			wantErr: "nats nkey",
		},
		{
			name: "a creds file",
			conn: config.NATSConnection{
				Host: "hub",
				Port: 4222,
				Auth: config.NATSAuth{Type: "creds", CredsFile: "/pki/nats.creds"},
			},
			wantURL:  "nats://hub:4222",
			wantOpts: 3,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			url, opts, err := agent.ExportNATSDialOptions(tt.conn)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tt.wantURL, url)
			s.Len(opts, tt.wantOpts)
		})
	}
}
