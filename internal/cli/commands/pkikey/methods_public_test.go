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

package pkikey_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/cli/commands/pkikey"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite
	fs     avfs.VFS
	kv     *mocks.MockKV
	audits []audit.Entry
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.fs = memfs.New()
	s.kv = mocks.NewMockKV(gomock.NewController(s.T()))
	s.audits = nil
}

// deps is a host with each kind of key in the config: the controller's
// and the agent's in their key directories, a JetStream key file, a vault
// with a key file, one with a key in a variable, one with neither, and a
// vault of another type, which has no key of osapi.
func (s *MethodsPublicTestSuite) deps() pkikey.Deps {
	var c config.Config
	c.Controller.PKI.KeyDir = "/ctl"
	c.Agent.PKI.KeyDir = "/agt"
	c.NATS.Server.JetStream.KeyFile = "/nats/js.key"
	c.NATS.Secrets.Bucket = "secrets"
	c.Controller.Secrets.Vaults = []vault.Spec{
		{Name: "default", Type: natsvault.TypeName, KeyFile: "/v/default.key"},
		{Name: "env", Type: natsvault.TypeName, KeyEnv: "VAULT_KEY"},
		{Name: "none", Type: natsvault.TypeName},
		{Name: "other", Type: "aws"},
	}

	return pkikey.Deps{
		FS:     s.fs,
		Config: c,
		Identity: func() (*identity.Identity, error) {
			return &identity.Identity{MachineID: "m-1", Hostname: "web-01"}, nil
		},
		Dial: func(context.Context) (natsvault.KV, func(), error) { return s.kv, func() {}, nil },
		LookupEnv: func(k string) (string, bool) {
			if k == "VAULT_KEY" {
				return strings.Repeat("ab", 32), true
			}

			return "", false
		},
		Audit: func(_ context.Context, e audit.Entry) error {
			s.audits = append(s.audits, e)

			return nil
		},
	}
}

// failing is fs, where the call fn fails.
func failing(
	fs avfs.VFS,
	fn avfs.FnVFS,
) avfs.VFS {
	vfs := failfs.New(fs)
	_ = vfs.SetFailFunc(func(_ avfs.VFSBase, got avfs.FnVFS, _ *failfs.FailParam) error {
		if got == fn {
			return errors.New("injected")
		}

		return nil
	})

	return vfs
}

// byName is the keys of list by name.
func byName(
	keys []pkikey.Key,
) map[string]pkikey.Key {
	out := map[string]pkikey.Key{}
	for _, k := range keys {
		out[k.Name] = k
	}

	return out
}

func (s *MethodsPublicTestSuite) TestList() {
	s.Run("a host where nothing has started has no key but the one in a variable", func() {
		got, err := pkikey.List(context.Background(), s.deps(), pkikey.ListInput{})
		s.Require().NoError(err)
		keys := byName(got.Items)
		s.Len(keys, 7, "the vault of another type has no key of osapi")
		for _, n := range []string{"controller", "agent", "agent-encryption", "jetstream", "vault:default", "vault:none"} {
			s.Equal(pkikey.StateNotMade, keys[n].State, n)
		}
		s.Equal(pkikey.StateOK, keys["vault:env"].State)
		s.Equal("$VAULT_KEY", keys["vault:env"].File)
		s.Len(keys["vault:env"].KeyID, 16)
	})

	s.Run("each key that osapi made, with its fingerprint or its ID, never the key", func() {
		d := s.deps()
		for _, n := range []string{"controller", "agent", "jetstream", "vault:default"} {
			_, err := pkikey.Create(context.Background(), d, pkikey.CreateInput{Name: n})
			s.Require().NoError(err, n)
		}
		got, _ := pkikey.List(context.Background(), d, pkikey.ListInput{})
		keys := byName(got.Items)
		ctl := pki.New(s.fs, "/ctl", "controller")
		s.Require().NoError(ctl.LoadOrGenerate())
		s.Equal(ctl.Fingerprint(), keys["controller"].Fingerprint)
		s.Equal("/ctl/controller.key", keys["controller"].File)
		agt := pki.New(s.fs, "/agt", "agent", pki.WithEncryptionKey())
		s.Require().NoError(agt.LoadOrGenerate())
		s.Equal(
			agt.Fingerprint(),
			keys["agent"].Fingerprint,
			"it covers both keys, as the admission does",
		)
		s.Equal(agt.EncryptionPublicKey(), keys["agent-encryption"].PublicKey)
		s.Len(keys["jetstream"].KeyID, 16)
		s.Equal("AES-256-GCM", keys["jetstream"].Algorithm, "the cipher that bus sets")
		s.Equal(1, keys["vault:default"].Keys)
		data, _ := s.fs.ReadFile("/nats/js.key")
		s.NotContains(keys["jetstream"].KeyID, strings.TrimSpace(string(data))[:16])
	})

	s.Run("a JetStream store with no key_file is off", func() {
		d := s.deps()
		d.Config.NATS.Server.JetStream.KeyFile = ""
		keys := byName(pkikey.Keys(d))
		s.Equal(pkikey.StateOff, keys["jetstream"].State)
		s.Empty(keys["jetstream"].File)
	})

	s.Run("each key that cannot be read says why", func() {
		s.Require().NoError(s.fs.MkdirAll("/ctl", 0o700))
		s.Require().NoError(s.fs.WriteFile("/ctl/controller.pub", []byte("x"), 0o644))
		s.Require().NoError(s.fs.MkdirAll("/agt", 0o700))
		s.Require().NoError(s.fs.WriteFile("/agt/agent.key", []byte("x"), 0o600))
		s.Require().NoError(s.fs.WriteFile("/agt/agent.ekey", []byte("not a key"), 0o600))
		s.Require().NoError(s.fs.MkdirAll("/nats", 0o700))
		s.Require().NoError(s.fs.WriteFile("/nats/js.key", []byte("  \n"), 0o600))
		s.Require().NoError(s.fs.MkdirAll("/v", 0o700))
		s.Require().NoError(s.fs.WriteFile("/v/default.key", []byte("short\n"), 0o600))
		d := s.deps()
		d.LookupEnv = func(string) (string, bool) { return "", false }
		keys := byName(pkikey.Keys(d))
		s.Equal(
			"unreadable: half a key pair: /ctl/controller.key is not there",
			keys["controller"].State,
		)
		s.Equal("unreadable: half a key pair: /agt/agent.pub is not there", keys["agent"].State)
		s.Contains(keys["agent-encryption"].State, "unreadable: /agt/agent.ekey: ")
		s.Equal("unreadable: /nats/js.key is empty", keys["jetstream"].State)
		s.Contains(keys["vault:default"].State, "unreadable: line 1")
		s.Equal(pkikey.StateNotMade, keys["vault:env"].State)
	})

	s.Run("a private key that cannot be read", func() {
		d := s.deps()
		d.FS = failing(s.fs, avfs.FnStat)
		s.Equal("unreadable: injected", byName(pkikey.Keys(d))["controller"].State)
	})

	s.Run("a public key that is not PEM, or a file that cannot be read", func() {
		s.Require().NoError(s.fs.MkdirAll("/ctl", 0o700))
		s.Require().NoError(s.fs.WriteFile("/ctl/controller.key", []byte("x"), 0o600))
		s.Require().NoError(s.fs.WriteFile("/ctl/controller.pub", []byte("x"), 0o644))
		s.Require().NoError(s.fs.MkdirAll("/agt/agent.key", 0o700))
		s.Require().NoError(s.fs.WriteFile("/agt/agent.pub", []byte("x"), 0o644))
		keys := byName(pkikey.Keys(s.deps()))
		s.Contains(keys["controller"].State, "unreadable: /ctl/controller.pub: ")
		s.NotEqual(pkikey.StateNotMade, keys["agent"].State, "a key that is a directory")
		s.Contains(keys["agent"].State, "unreadable")
	})
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name     string
		in       string
		make     []string
		identity func() (*identity.Identity, error)
		check    func(pkikey.Key)
		wantErr  string
	}{
		{
			name: "the agent's key has the command that admits this machine",
			in:   "agent", make: []string{"agent"},
			check: func(k pkikey.Key) {
				s.Contains(
					k.Admit,
					"osapi client admission admit --machine-id m-1 --fingerprint "+k.Fingerprint,
				)
			},
		},
		{
			name:     "the agent's key with no identity has no admit command",
			in:       "agent",
			make:     []string{"agent"},
			identity: func() (*identity.Identity, error) { return nil, errors.New("no machine-id") },
			check:    func(k pkikey.Key) { s.Empty(k.Admit) },
		},
		{
			name: "another key is as list shows it",
			in:   "controller", make: []string{"controller"},
			check: func(k pkikey.Key) {
				s.Equal(pkikey.StateOK, k.State)
				s.Empty(k.Admit)
			},
		},
		{
			name:    "a key that is not on this host names create",
			in:      "jetstream",
			wantErr: "jetstream is not on this host: nats; make it with osapi pki key create --name jetstream: not made",
		},
		{
			name:    "a name that no key has",
			in:      "nope",
			wantErr: "nope: no key with this name: list shows each",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := s.deps()
			if tc.identity != nil {
				d.Identity = tc.identity
			}
			for _, n := range tc.make {
				_, err := pkikey.Create(context.Background(), d, pkikey.CreateInput{Name: n})
				s.Require().NoError(err)
			}
			got, err := pkikey.Get(context.Background(), d, pkikey.GetInput{Name: tc.in})
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.check(got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestCreate() {
	tests := []struct {
		name    string
		in      string
		setup   func(d *pkikey.Deps)
		wantErr string
		check   func(pkikey.Key)
	}{
		{
			name: "the JetStream key, audited by its ID",
			in:   "jetstream",
			check: func(k pkikey.Key) {
				s.Equal(pkikey.StateOK, k.State)
				st, err := s.fs.Stat("/nats/js.key")
				s.Require().NoError(err)
				s.Equal("-rw-------", st.Mode().Perm().String())
				s.Require().Len(s.audits, 1)
				s.Equal("key.create", s.audits[0].Action)
				s.Equal("key_id="+k.KeyID, s.audits[0].RequestSummary)
			},
		},
		{
			name: "the two keys of the agent, by the name of either",
			in:   "agent-encryption",
			check: func(k pkikey.Key) {
				s.Equal("agent", k.Name)
				s.Equal(pkikey.StateOK, k.State)
				s.Equal("fingerprint="+k.Fingerprint, s.audits[0].RequestSummary)
			},
		},
		{
			name: "the encryption key of an agent that has only its signing key",
			in:   "agent",
			setup: func(*pkikey.Deps) {
				s.Require().NoError(pki.New(s.fs, "/agt", "agent").LoadOrGenerate())
			},
			check: func(pkikey.Key) {
				x := byName(pkikey.Keys(s.deps()))["agent-encryption"]
				s.Equal(pkikey.StateOK, x.State)
			},
		},
		{
			name:  "a vault key",
			in:    "vault:default",
			check: func(k pkikey.Key) { s.Equal(1, k.Keys) },
		},
		{
			name: "a key that is there is refused, with what to do instead",
			in:   "jetstream",
			setup: func(d *pkikey.Deps) {
				_, err := pkikey.Create(
					context.Background(),
					*d,
					pkikey.CreateInput{Name: "jetstream"},
				)
				s.Require().NoError(err)
			},
			wantErr: "jetstream: the key exists: a new key makes each stream and bucket",
		},
		{
			name: "a vault key that is there names rotate",
			in:   "vault:default",
			setup: func(*pkikey.Deps) {
				_, err := natsvault.CreateKey(context.Background(), s.fs, "/v/default.key")
				s.Require().NoError(err)
			},
			wantErr: "vault:default: the key exists: osapi pki key rotate --name vault:default makes a new one",
		},
		{
			name: "a key that cannot be read is not replaced",
			in:   "jetstream",
			setup: func(*pkikey.Deps) {
				s.Require().NoError(s.fs.MkdirAll("/nats", 0o700))
				s.Require().NoError(s.fs.WriteFile("/nats/js.key", []byte("\n"), 0o600))
			},
			wantErr: "jetstream is unreadable: /nats/js.key is empty: osapi does not replace a key that it cannot read",
		},
		{
			name:    "a JetStream store with no key_file",
			in:      "jetstream",
			setup:   func(d *pkikey.Deps) { d.Config.NATS.Server.JetStream.KeyFile = "" },
			wantErr: "nats.server.jetstream.key_file is not set",
		},
		{
			name: "random bytes that cannot be read",
			in:   "jetstream",
			setup: func(*pkikey.Deps) {
				s.T().
					Cleanup(pkikey.SetRandRead(func([]byte) (int, error) { return 0, errors.New("no entropy") }))
			},
			wantErr: "make the key: no entropy",
		},
		{
			name:    "a JetStream key file that cannot be written",
			in:      "jetstream",
			setup:   func(d *pkikey.Deps) { d.FS = failing(s.fs, avfs.FnRename) },
			wantErr: "write /nats/js.key: ",
		},
		{
			name:    "a vault with no key_file",
			in:      "vault:none",
			wantErr: `vault "none" has no key_file: set it to the file to write, or put the key in its key_env`,
		},
		{
			name:    "a vault key file that cannot be written",
			in:      "vault:default",
			setup:   func(d *pkikey.Deps) { d.FS = failing(s.fs, avfs.FnMkdirAll) },
			wantErr: "vault:default: make the directory of /v/default.key: ",
		},
		{
			name: "an agent key that pki cannot load",
			in:   "agent",
			setup: func(*pkikey.Deps) {
				s.Require().NoError(pki.New(s.fs, "/agt", "agent").LoadOrGenerate())
				s.Require().NoError(s.fs.WriteFile("/agt/agent.key", []byte("not PEM"), 0o600))
			},
			wantErr: "make the agent key in /agt: ",
		},
		{name: "a name that no key has", in: "nope", wantErr: "nope: no key with this name"},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := s.deps()
			if tc.setup != nil {
				tc.setup(&d)
			}
			got, err := pkikey.Create(context.Background(), d, pkikey.CreateInput{Name: tc.in})
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.check(got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestRotate() {
	tests := []struct {
		name    string
		in      string
		setup   func(d *pkikey.Deps)
		wantErr string
	}{
		{
			name: "a vault key, and each value again with it",
			in:   "vault:default",
			setup: func(*pkikey.Deps) {
				s.kv.EXPECT().KVStatus(gomock.Any(), "secrets").Return(nil, nil)
				s.kv.EXPECT().KVKeys(gomock.Any(), "secrets", "default.>").Return(nil, nil)
			},
		},
		{
			name: "the vault of no bucket",
			in:   "vault:default",
			setup: func(d *pkikey.Deps) {
				d.Config.NATS.Secrets.Bucket = ""
			},
			wantErr: `vault "default": nats.secrets.bucket is empty`,
		},
		{
			name: "a rotation that fails",
			in:   "vault:default",
			setup: func(*pkikey.Deps) {
				s.kv.EXPECT().KVStatus(gomock.Any(), "secrets").Return(nil, nil)
				s.kv.EXPECT().KVKeys(gomock.Any(), "secrets", "default.>").
					Return(nil, errors.New("timeout"))
			},
			wantErr: `vault "default": list default: timeout`,
		},
		{
			name: "a bucket that is not there",
			in:   "vault:default",
			setup: func(*pkikey.Deps) {
				s.kv.EXPECT().
					KVStatus(gomock.Any(), "secrets").
					Return(nil, jetstream.ErrBucketNotFound)
			},
			wantErr: `vault "default": open the bucket secrets: `,
		},
		{
			name: "NATS that does not answer",
			in:   "vault:default",
			setup: func(d *pkikey.Deps) {
				d.Dial = func(context.Context) (natsvault.KV, func(), error) {
					return nil, nil, errors.New("no route")
				}
			},
			wantErr: "no route",
		},
		{
			name:    "the JetStream key is refused",
			in:      "jetstream",
			wantErr: "jetstream: osapi does not rotate this key: a new key makes each stream",
		},
		{
			name:    "the controller's key is refused",
			in:      "controller",
			wantErr: "controller: osapi does not rotate this key: each admitted agent verifies",
		},
		{
			name:    "the agent's keys are refused",
			in:      "agent-encryption",
			wantErr: "agent-encryption: osapi does not rotate this key: the controller admitted",
		},
		{name: "a name that no key has", in: "nope", wantErr: "nope: no key with this name"},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := s.deps()
			_, err := natsvault.CreateKey(context.Background(), s.fs, "/v/default.key")
			s.Require().NoError(err)
			if tc.setup != nil {
				tc.setup(&d)
			}
			got, err := pkikey.Rotate(context.Background(), d, pkikey.RotateInput{Name: tc.in})
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(1, got.Key.Keys)
			s.Equal(0, got.Values)
			s.Require().Len(s.audits, 1)
			s.Equal("key.rotate", s.audits[0].Action)
		})
	}
}
