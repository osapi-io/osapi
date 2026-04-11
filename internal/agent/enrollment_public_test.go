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
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"log/slog"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/config"
)

type EnrollmentPublicTestSuite struct {
	suite.Suite
}

func (suite *EnrollmentPublicTestSuite) TestHandlePKIEnrollment() {
	tests := []struct {
		name         string
		setupFS      func() avfs.VFS
		wantErr      bool
		wantContains string
		validateFunc func(a *agent.Agent)
	}{
		{
			name: "when no existing keys generates keypair and enters pending state",
			setupFS: func() avfs.VFS {
				return memfs.New()
			},
			validateFunc: func(a *agent.Agent) {
				m := agent.GetAgentPKIManager(a)
				require.NotNil(suite.T(), m)
				assert.Len(suite.T(), m.PublicKey(), ed25519.PublicKeySize)
				assert.Len(suite.T(), m.PrivateKey(), ed25519.PrivateKeySize)
				assert.NotEmpty(suite.T(), m.Fingerprint())
				// No controller key — not enrolled.
				assert.Nil(suite.T(), m.ControllerPublicKey())
			},
		},
		{
			name: "when controller public key exists loads and marks enrolled",
			setupFS: func() avfs.VFS {
				fs := memfs.New()
				_ = fs.MkdirAll("/keys", 0o700)

				// Pre-generate agent keys.
				pub, priv, _ := ed25519.GenerateKey(rand.Reader)
				privPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "ED25519 PRIVATE KEY",
					Bytes: priv.Seed(),
				})
				pubPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "ED25519 PUBLIC KEY",
					Bytes: pub,
				})
				_ = fs.WriteFile("/keys/agent.key", privPEM, 0o600)
				_ = fs.WriteFile("/keys/agent.pub", pubPEM, 0o644)

				// Write controller public key.
				ctrlPub, _, _ := ed25519.GenerateKey(rand.Reader)
				ctrlPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "ED25519 PUBLIC KEY",
					Bytes: ctrlPub,
				})
				_ = fs.WriteFile("/keys/controller.pub", ctrlPEM, 0o644)

				return fs
			},
			validateFunc: func(a *agent.Agent) {
				m := agent.GetAgentPKIManager(a)
				require.NotNil(suite.T(), m)
				assert.Len(suite.T(), m.PublicKey(), ed25519.PublicKeySize)
				assert.Len(suite.T(), m.ControllerPublicKey(), ed25519.PublicKeySize)
			},
		},
		{
			name: "when controller public key PEM is corrupted returns error",
			setupFS: func() avfs.VFS {
				fs := memfs.New()
				_ = fs.MkdirAll("/keys", 0o700)

				// Pre-generate agent keys.
				pub, priv, _ := ed25519.GenerateKey(rand.Reader)
				privPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "ED25519 PRIVATE KEY",
					Bytes: priv.Seed(),
				})
				pubPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "ED25519 PUBLIC KEY",
					Bytes: pub,
				})
				_ = fs.WriteFile("/keys/agent.key", privPEM, 0o600)
				_ = fs.WriteFile("/keys/agent.pub", pubPEM, 0o644)

				// Write corrupted controller public key.
				_ = fs.WriteFile("/keys/controller.pub", []byte("not valid pem"), 0o644)

				return fs
			},
			wantErr:      true,
			wantContains: "failed to parse controller public key",
		},
		{
			name: "when controller public key PEM has wrong block type returns error",
			setupFS: func() avfs.VFS {
				fs := memfs.New()
				_ = fs.MkdirAll("/keys", 0o700)

				// Pre-generate agent keys.
				pub, priv, _ := ed25519.GenerateKey(rand.Reader)
				privPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "ED25519 PRIVATE KEY",
					Bytes: priv.Seed(),
				})
				pubPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "ED25519 PUBLIC KEY",
					Bytes: pub,
				})
				_ = fs.WriteFile("/keys/agent.key", privPEM, 0o600)
				_ = fs.WriteFile("/keys/agent.pub", pubPEM, 0o644)

				// Write controller key with wrong block type.
				ctrlPub, _, _ := ed25519.GenerateKey(rand.Reader)
				ctrlPEM := pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PUBLIC KEY",
					Bytes: ctrlPub,
				})
				_ = fs.WriteFile("/keys/controller.pub", ctrlPEM, 0o644)

				return fs
			},
			wantErr:      true,
			wantContains: "failed to parse controller public key",
		},
		{
			name: "when keypair generation fails returns error",
			setupFS: func() avfs.VFS {
				vfs := failfs.New(memfs.New())
				_ = vfs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					_ *failfs.FailParam,
				) error {
					if fn == avfs.FnMkdirAll {
						return errors.New("permission denied")
					}
					return nil
				})

				return vfs
			},
			wantErr:      true,
			wantContains: "create key directory",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			fs := tc.setupFS()

			logger := slog.Default()
			cfg := config.Config{
				Agent: config.AgentConfig{
					PKI: config.AgentPKI{
						Enabled: true,
						KeyDir:  "/keys",
					},
				},
			}

			a := agent.New(
				fs,
				cfg,
				logger,
				nil, // jobClient
				"",  // streamName
				nil, // hostProvider
				nil, // diskProvider
				nil, // memProvider
				nil, // loadProvider
				nil, // netinfoProvider
				nil, // processProvider
				nil, // registry
				nil, // registryKV
				nil, // factsKV
				nil, // execManager
			)

			err := agent.ExportHandlePKIEnrollment(context.Background(), a)

			if tc.wantErr {
				require.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.wantContains)
			} else {
				require.NoError(suite.T(), err)
				if tc.validateFunc != nil {
					tc.validateFunc(a)
				}
			}
		})
	}
}

func TestEnrollmentPublicTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(EnrollmentPublicTestSuite))
}
