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

package secret_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/endpoints/secret"
	"github.com/osapi-io/osapi/internal/controller/endpoints/secret/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoints/secret/mocks"
	"github.com/osapi-io/osapi/internal/secret/vault"
	vaultmocks "github.com/osapi-io/osapi/internal/secret/vault/mocks"
)

var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-secret-group"}

var (
	errBoom = errors.New("nats: timeout")
	changed = time.Date(2026, 1, 15, 9, 12, 0, 0, time.UTC)
	meta    = vault.Meta{Key: "registry-token", Version: 3, Changed: changed, By: "ops@example.com"}
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	vaults *mocks.MockVaults
	vault  *vaultmocks.MockVault
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.vaults = mocks.NewMockVaults(s.ctrl)
	s.vault = vaultmocks.NewMockVault(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) deps() secret.Deps {
	return secret.Deps{Vaults: s.vaults}
}

// hasDefault makes the vault default the one that the mock gives.
func (s *MethodsPublicTestSuite) hasDefault() {
	s.vaults.EXPECT().Get("default").Return(s.vault, true)
}

func (s *MethodsPublicTestSuite) TestNoVault() {
	ctx := context.Background()
	d := secret.Deps{}
	s.Run("when the controller has no vaults, each method is ErrNoVault", func() {
		for _, call := range []func() error{
			func() error { _, err := secret.List(ctx, d, secret.ListInput{Vault: "default"}); return err },
			func() error {
				_, err := secret.Get(ctx, d, secret.GetInput{Vault: "default", Key: "k"})
				return err
			},
			func() error {
				_, err := secret.Update(ctx, d, secret.UpdateInput{Vault: "default", Key: "k", Body: secret.SecretUpdateRequest{Value: "v"}})
				return err
			},
			func() error {
				_, err := secret.Delete(ctx, d, secret.DeleteInput{Vault: "default", Key: "k"})
				return err
			},
		} {
			err := call()
			s.ErrorIs(err, secret.ErrNoVault)
			s.ErrorContains(err, "no vault named default")
		}
	})
	s.Run("when the set has no vault with the name, it is ErrNoVault", func() {
		s.vaults.EXPECT().Get("nope").Return(nil, false)
		_, err := secret.List(ctx, s.deps(), secret.ListInput{Vault: "nope"})
		s.ErrorIs(err, secret.ErrNoVault)
	})
}

func (s *MethodsPublicTestSuite) TestVaultList() {
	tests := []struct {
		name  string
		deps  func() secret.Deps
		setup func()
		want  []secret.VaultInfo
	}{
		{
			name: "when the controller has no vaults, the list is empty",
			deps: func() secret.Deps { return secret.Deps{} },
			want: []secret.VaultInfo{},
		},
		{
			name: "when a vault answers and one does not, each says so",
			deps: s.deps,
			setup: func() {
				other := vaultmocks.NewMockVault(s.ctrl)
				s.vaults.EXPECT().Names().Return([]string{"default", "prod"})
				s.vaults.EXPECT().
					Spec("default").
					Return(vault.Spec{Name: "default", Type: "nats"}, true)
				s.vaults.EXPECT().Spec("prod").
					Return(vault.Spec{Name: "prod", Type: "nats", Targets: []string{"group:prod"}}, true)
				s.vaults.EXPECT().Get("default").Return(s.vault, true)
				s.vaults.EXPECT().Get("prod").Return(other, true)
				s.vault.EXPECT().List(gomock.Any()).Return(nil, nil)
				other.EXPECT().List(gomock.Any()).Return(nil, errBoom)
			},
			want: []secret.VaultInfo{
				{Name: "default", Type: "nats", Targets: []string{}, Answers: true},
				{Name: "prod", Type: "nats", Targets: []string{"group:prod"}},
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}
			got, err := secret.VaultList(context.Background(), tc.deps(), secret.VaultListInput{})
			s.Require().NoError(err)
			s.Equal(tc.want, got.Items)
			s.Equal(len(tc.want), got.TotalItems)
		})
	}
}

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name    string
		setup   func()
		want    []secret.SecretInfo
		wantErr string
	}{
		{
			name: "when the vault has keys, the list has the metadata of each",
			setup: func() {
				s.vault.EXPECT().List(gomock.Any()).Return([]vault.Meta{meta}, nil)
			},
			want: []secret.SecretInfo{
				{
					Vault:   "default",
					Key:     "registry-token",
					Version: 3,
					Changed: changed,
					By:      "ops@example.com",
				},
			},
		},
		{
			name: "when the vault has no keys, the list is empty",
			setup: func() {
				s.vault.EXPECT().List(gomock.Any()).Return(nil, nil)
			},
			want: []secret.SecretInfo{},
		},
		{
			name: "when the vault fails, it is a failure of the controller",
			setup: func() {
				s.vault.EXPECT().List(gomock.Any()).Return(nil, errBoom)
			},
			wantErr: "failed to list the keys of the vault",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.hasDefault()
			tc.setup()
			got, err := secret.List(
				context.Background(),
				s.deps(),
				secret.ListInput{Vault: "default"},
			)
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got.Items)
			s.Equal(len(tc.want), got.TotalItems)
		})
	}
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name    string
		key     string
		setup   func()
		wantErr error
		wantMsg string
	}{
		{
			name: "when the key is there, the answer is its metadata, from the list",
			key:  "registry-token",
			setup: func() {
				s.vault.EXPECT().List(gomock.Any()).
					Return([]vault.Meta{{Key: "db-password"}, meta}, nil)
			},
		},
		{
			name: "when the key is not there, it is ErrNoKey",
			key:  "nope",
			setup: func() {
				s.vault.EXPECT().List(gomock.Any()).Return([]vault.Meta{meta}, nil)
			},
			wantErr: secret.ErrNoKey,
			wantMsg: "vault default has no key nope",
		},
		{
			name: "when the vault fails, it is a failure of the controller",
			key:  "registry-token",
			setup: func() {
				s.vault.EXPECT().List(gomock.Any()).Return(nil, errBoom)
			},
			wantMsg: "failed to list the keys of the vault",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.hasDefault()
			tc.setup()
			got, err := secret.Get(context.Background(), s.deps(),
				secret.GetInput{Vault: "default", Key: tc.key})
			if tc.wantMsg != "" {
				s.ErrorContains(err, tc.wantMsg)
				if tc.wantErr != nil {
					s.ErrorIs(err, tc.wantErr)
				}

				return
			}
			s.Require().NoError(err)
			s.Equal(secret.SecretInfo{
				Vault:   "default",
				Key:     "registry-token",
				Version: 3,
				Changed: changed,
				By:      "ops@example.com",
			}, got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestUpdate() {
	tests := []struct {
		name    string
		setup   func()
		wantErr string
	}{
		{
			name: "when the vault writes the value, the answer is the metadata, with the subject as who",
			setup: func() {
				s.vault.EXPECT().Put(gomock.Any(), "registry-token", []byte("s3cret")).
					DoAndReturn(func(ctx context.Context, key string, _ []byte) (vault.Meta, error) {
						s.Equal("ops@example.com", vault.Actor(ctx), "the vault records who")

						return vault.Meta{
							Key:     key,
							Version: 4,
							Changed: changed,
							By:      vault.Actor(ctx),
						}, nil
					})
			},
		},
		{
			name: "when the vault fails, it is a failure of the controller",
			setup: func() {
				s.vault.EXPECT().Put(gomock.Any(), "registry-token", gomock.Any()).
					Return(vault.Meta{}, errBoom)
			},
			wantErr: "failed to write the key",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.hasDefault()
			tc.setup()
			ctx := authtoken.WithSubject(context.Background(), "ops@example.com")
			got, err := secret.Update(
				ctx,
				s.deps(),
				secret.UpdateInput{
					Vault: "default",
					Key:   "registry-token",
					Body:  secret.SecretUpdateRequest{Value: "s3cret"},
				},
			)
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.NotContains(err.Error(), "s3cret")

				return
			}
			s.Require().NoError(err)
			s.Equal(secret.SecretInfo{
				Vault:   "default",
				Key:     "registry-token",
				Version: 4,
				Changed: changed,
				By:      "ops@example.com",
			}, got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestDelete() {
	tests := []struct {
		name    string
		err     error
		wantErr error
		wantMsg string
	}{
		{name: "when the key is there, it is deleted"},
		{
			name:    "when the key is not there, it is ErrNoKey",
			err:     vault.ErrNotFound,
			wantErr: secret.ErrNoKey,
			wantMsg: "vault default has no key registry-token",
		},
		{
			name:    "when the vault fails, it is a failure of the controller",
			err:     errBoom,
			wantMsg: "failed to delete the key",
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.hasDefault()
			s.vault.EXPECT().Delete(gomock.Any(), "registry-token").Return(tc.err)
			_, err := secret.Delete(context.Background(), s.deps(),
				secret.DeleteInput{Vault: "default", Key: "registry-token"})
			if tc.wantMsg != "" {
				s.ErrorContains(err, tc.wantMsg)
				if tc.wantErr != nil {
					s.ErrorIs(err, tc.wantErr)
				}

				return
			}
			s.Require().NoError(err)
		})
	}
}

func (s *MethodsPublicTestSuite) TestRoutes() {
	tests := []struct {
		name     string
		method   string
		path     string
		body     string
		token    string
		setup    func()
		wantCode int
		wantBody string
		notBody  string
	}{
		{
			name:   "when a caller with secret:read lists the vaults, it is a 200",
			method: http.MethodGet,
			path:   "/api/secret",
			token:  token("secret:read"),
			setup: func() {
				s.vaults.EXPECT().Names().Return(nil)
			},
			wantCode: http.StatusOK,
			wantBody: `"total_items":0`,
		},
		{
			name:   "when a caller gets a key, the answer is its metadata",
			method: http.MethodGet,
			path:   "/api/secret/default/registry-token",
			token:  token("secret:read"),
			setup: func() {
				s.hasDefault()
				s.vault.EXPECT().List(gomock.Any()).Return([]vault.Meta{meta}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `"version":3`,
		},
		{
			name:   "when a key is not there, it is a 404",
			method: http.MethodGet,
			path:   "/api/secret/default/nope",
			token:  token("secret:read"),
			setup: func() {
				s.hasDefault()
				s.vault.EXPECT().List(gomock.Any()).Return(nil, nil)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:   "when a vault is not there, it is a 404",
			method: http.MethodGet,
			path:   "/api/secret/nope",
			token:  token("secret:read"),
			setup: func() {
				s.vaults.EXPECT().Get("nope").Return(nil, false)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:   "when a caller with secret:write writes a value, the answer has no value",
			method: http.MethodPut,
			path:   "/api/secret/default/registry-token",
			body:   `{"value":"s3cret"}`,
			token:  token("secret:write"),
			setup: func() {
				s.hasDefault()
				s.vault.EXPECT().Put(gomock.Any(), "registry-token", []byte("s3cret")).
					Return(vault.Meta{Key: "registry-token", Version: 1, Changed: changed, By: "test"}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `"by":"test"`,
			notBody:  "s3cret",
		},
		{
			name:     "when a value is a reference, it is a 400 and nothing is written",
			method:   http.MethodPut,
			path:     "/api/secret/default/registry-token",
			body:     `{"value":"secret://default/other"}`,
			token:    token("secret:write"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "when a caller with only secret:read writes, it is a 403",
			method:   http.MethodPut,
			path:     "/api/secret/default/registry-token",
			body:     `{"value":"s3cret"}`,
			token:    token("secret:read"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:   "when a caller deletes a key, it is a 204",
			method: http.MethodDelete,
			path:   "/api/secret/default/registry-token",
			token:  token("secret:write"),
			setup: func() {
				s.hasDefault()
				s.vault.EXPECT().Delete(gomock.Any(), "registry-token").Return(nil)
			},
			wantCode: http.StatusNoContent,
		},
		{
			name:     "when there is no token, it is a 401",
			method:   http.MethodGet,
			path:     "/api/secret",
			setup:    func() {},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, s.deps())
			e := echo.New()
			srv.Mount(e, gen.Routes(srv))

			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.Contains(rec.Body.String(), tc.wantBody)
			}
			if tc.notBody != "" {
				s.NotContains(rec.Body.String(), tc.notBody)
			}
		})
	}
}

// token is a token with the observer role and perms, whose subject is
// test.
func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).
		Generate(signingKey, []string{"observer"}, "test", perms, nil)

	return t
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
