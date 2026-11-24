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
package file_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/endpoints/file"
	"github.com/osapi-io/osapi/internal/controller/endpoints/file/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoints/file/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-file-group"}

var (
	content = []byte("server { listen 80; }")
	errBoom = errors.New("nats: timeout")
)

// digest is the Object Store digest of data.
func digest(
	data []byte,
) string {
	sum := sha256.Sum256(data)

	return "SHA-256=" + base64.URLEncoding.EncodeToString(sum[:])
}

func hexOf(
	data []byte,
) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}

// entry is a KV entry with a value.
type entry struct {
	jetstream.KeyValueEntry
	value []byte
}

func (e entry) Value() []byte { return e.value }

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl    *gomock.Controller
	objects *mocks.MockObjects
	states  *mocks.MockStates
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.objects = mocks.NewMockObjects(s.ctrl)
	s.states = mocks.NewMockStates(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) deps() file.Deps {
	return file.Deps{Objects: s.objects, States: s.states}
}

// answered is the status an Answer error names, or 0.
func answered(
	err error,
) int {
	var a *endpoint.AnswerError
	if errors.As(err, &a) {
		return a.Code
	}

	return 0
}

// TestSpecContentType: the spec's ContentType has each content type, so a
// client checks a value when it compiles.
func (s *MethodsPublicTestSuite) TestSpecContentType() {
	doc, err := endpoint.OpenAPI(file.Group)
	s.Require().NoError(err)

	s.Equal(
		[]any{string(types.ContentRaw), string(types.ContentTemplate)},
		doc.Components.Schemas["ContentType"].Value.Enum,
	)
}

func (s *MethodsPublicTestSuite) TestNoStore() {
	ctx := context.Background()
	d := file.Deps{}
	for name, call := range map[string]func() error{
		"list":   func() error { _, err := file.List(ctx, d, file.ListInput{}); return err },
		"get":    func() error { _, err := file.Get(ctx, d, file.GetInput{Name: "a"}); return err },
		"upload": func() error { _, err := file.Upload(ctx, d, file.UploadInput{Name: "a"}); return err },
		"update": func() error { _, err := file.Update(ctx, d, file.UpdateInput{Name: "a"}); return err },
		"delete": func() error { _, err := file.Delete(ctx, d, file.DeleteInput{Name: "a"}); return err },
		"stale":  func() error { _, err := file.Stale(ctx, d, file.StaleInput{}); return err },
	} {
		s.Run("when there is no Object Store, "+name+" is ErrNoStore", func() {
			err := call()
			s.ErrorIs(err, file.ErrNoStore)
			s.ErrorContains(err, "nats.objects.bucket")
		})
	}
}

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func(file.FileListResponse, error)
	}{
		{
			name: "when the store has files, each one but a deleted one is in the list",
			setup: func() {
				s.objects.EXPECT().List(gomock.Any()).Return([]*jetstream.ObjectInfo{
					{
						ObjectMeta: jetstream.ObjectMeta{
							Name:    "nginx.conf",
							Headers: nats.Header{"Osapi-Content-Type": []string{"template"}},
						},
						Size: 21, Digest: digest(content),
					},
					{
						ObjectMeta: jetstream.ObjectMeta{Name: "osapi/cron.tmpl"},
						Digest:     "SHA-256=!!",
					},
					{ObjectMeta: jetstream.ObjectMeta{Name: "gone"}, Deleted: true},
				}, nil)
			},
			validateFunc: func(got file.FileListResponse, err error) {
				s.Require().NoError(err)
				s.Equal(2, got.TotalItems)
				s.Equal(file.FileInfo{
					Name: "nginx.conf", Sha256: hexOf(content), Size: 21,
					ContentType: "template", Source: "user",
				}, got.Items[0])
				s.Equal("osapi", got.Items[1].Source)
				s.Equal("!!", got.Items[1].Sha256)
				s.Empty(got.Items[1].ContentType)
			},
		},
		{
			name:  "when the store fails, the caller gets a message without its error",
			setup: func() { s.objects.EXPECT().List(gomock.Any()).Return(nil, errBoom) },
			validateFunc: func(_ file.FileListResponse, err error) {
				s.EqualError(err, "failed to list files")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			tc.validateFunc(file.List(context.Background(), s.deps(), file.ListInput{}))
		})
	}
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func(file.FileInfoResponse, error)
	}{
		{
			name: "when the file is there, it answers its metadata",
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").Return(&jetstream.ObjectInfo{
					ObjectMeta: jetstream.ObjectMeta{
						Name: "nginx.conf",
					}, Size: 21, Digest: digest(content),
				}, nil)
			},
			validateFunc: func(got file.FileInfoResponse, err error) {
				s.Require().NoError(err)
				s.Equal(
					file.FileInfoResponse{Name: "nginx.conf", Sha256: hexOf(content), Size: 21},
					got,
				)
			},
		},
		{
			name: "when the file is not there, it is ErrNotFound",
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(nil, jetstream.ErrObjectNotFound)
			},
			validateFunc: func(_ file.FileInfoResponse, err error) { s.ErrorIs(err, file.ErrNotFound) },
		},
		{
			name: "when the store fails, it is a 500",
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").Return(nil, errBoom)
			},
			validateFunc: func(_ file.FileInfoResponse, err error) {
				s.EqualError(err, "failed to get file info")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			tc.validateFunc(
				file.Get(context.Background(), s.deps(), file.GetInput{Name: "nginx.conf"}),
			)
		})
	}
}

func (s *MethodsPublicTestSuite) TestUploadAndUpdate() {
	exists := &jetstream.ObjectInfo{Digest: digest(content)}
	other := &jetstream.ObjectInfo{Digest: "SHA-256=old"}

	tests := []struct {
		name     string
		update   bool
		fileName string
		setup    func()
		wantErr  error
		wantCode int
		changed  bool
	}{
		{
			name: "when an upload names a new file, it is stored: 201", fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(nil, jetstream.ErrObjectNotFound)
				s.objects.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{}, nil)
			},
			changed: true,
		},
		{
			name:     "when an upload has the same content, nothing is written: 200",
			fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").Return(exists, nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name: "when an upload has other content, it is ErrExists", fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").Return(other, nil)
			},
			wantErr: file.ErrExists,
		},
		{
			name: "when an upload names an osapi file, it is ErrProtected", fileName: "osapi/x",
			setup: func() {}, wantErr: file.ErrProtected,
		},
		{
			name: "when the store fails, the upload is a 500", fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(nil, jetstream.ErrObjectNotFound)
				s.objects.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errBoom)
			},
			wantErr: errBoom,
		},
		{
			name:     "when an update names a new file, it is stored: 201",
			update:   true,
			fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(nil, jetstream.ErrObjectNotFound)
				s.objects.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{}, nil)
			},
			changed: true,
		},
		{
			name:     "when an update has other content, it is replaced: 200",
			update:   true,
			fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").Return(other, nil)
				s.objects.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{}, nil)
			},
			wantCode: http.StatusOK,
			changed:  true,
		},
		{
			name:     "when an update has the same content, nothing is written: 200",
			update:   true,
			fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").Return(exists, nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "when an update names an osapi file, it is ErrProtected",
			update:   true,
			fileName: "osapi/x",
			setup:    func() {},
			wantErr:  file.ErrProtected,
		},
		{
			name: "when the store fails, the update is a 500", update: true, fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(nil, jetstream.ErrObjectNotFound)
				s.objects.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errBoom)
			},
			wantErr: errBoom,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			var got file.FileUploadResponse
			var err error
			if tc.update {
				got, err = file.Update(context.Background(), s.deps(),
					file.UpdateInput{Name: tc.fileName, File: content})
			} else {
				got, err = file.Upload(context.Background(), s.deps(),
					file.UploadInput{Name: tc.fileName, File: content, ContentType: "raw"})
			}
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)

				return
			}
			s.Equal(tc.wantCode, answered(err))
			s.Equal(tc.changed, got.Changed)
			s.Equal(types.ContentRaw, got.ContentType)
			s.Equal(hexOf(content), got.Sha256)
		})
	}
}

func (s *MethodsPublicTestSuite) TestDelete() {
	tests := []struct {
		name     string
		fileName string
		setup    func()
		wantErr  error
	}{
		{
			name: "when the file is there, it goes", fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(&jetstream.ObjectInfo{}, nil)
				s.objects.EXPECT().Delete(gomock.Any(), "nginx.conf").Return(nil)
			},
		},
		{
			name: "when the file is an osapi file, it is ErrProtected", fileName: "osapi/x",
			setup: func() {}, wantErr: file.ErrProtected,
		},
		{
			name: "when the file is not there, it is ErrNotFound", fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(nil, jetstream.ErrObjectNotFound)
			},
			wantErr: file.ErrNotFound,
		},
		{
			name: "when the info fails, it is a 500", fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").Return(nil, errBoom)
			},
			wantErr: errBoom,
		},
		{
			name: "when the delete fails, it is a 500", fileName: "nginx.conf",
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(&jetstream.ObjectInfo{}, nil)
				s.objects.EXPECT().Delete(gomock.Any(), "nginx.conf").Return(errBoom)
			},
			wantErr: errBoom,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			_, err := file.Delete(
				context.Background(),
				s.deps(),
				file.DeleteInput{Name: tc.fileName},
			)
			if tc.wantErr == nil {
				s.NoError(err)

				return
			}
			s.ErrorIs(err, tc.wantErr)
		})
	}
}

func (s *MethodsPublicTestSuite) TestStale() {
	key := func(host string) string { return host + "." + hexOf([]byte(host)) }
	// state is a record of a deploy of object, from an object whose checksum
	// was objectSHA. The file's checksum is never the object's, as for a
	// template.
	state := func(object, path, objectSHA string, undeployed bool) []byte {
		u := ""
		if undeployed {
			u = `,"undeployed_at":"2026-11-17T00:00:00Z"`
		}

		return []byte(`{"object_name":"` + object + `","path":"` + path +
			`","sha256":"rendered","object_sha256":"` + objectSHA +
			`","deployed_at":"2026-11-17T00:00:00Z"` + u + `}`)
	}

	tests := []struct {
		name         string
		noStates     bool
		setup        func()
		validateFunc func(file.StaleDeploymentsResponse, error)
	}{
		{
			name: "when deploys changed, went, or match, only the changed and the gone are stale",
			setup: func() {
				s.states.EXPECT().Keys(gomock.Any()).Return([]string{
					key("web-01"), key("web-02"), key("web-03"), key("web-04"), "bad", "short",
				}, nil)
				s.states.EXPECT().Get(gomock.Any(), key("web-01")).
					Return(entry{value: state("svc", "/etc/systemd/system/a.service", "old", false)}, nil)
				s.states.EXPECT().Get(gomock.Any(), key("web-02")).
					Return(entry{value: state("ca", "/usr/local/share/ca-certificates/a.crt", "x", false)}, nil)
				s.states.EXPECT().Get(gomock.Any(), key("web-03")).
					Return(entry{value: state("cron", "/etc/cron.d/a", hexOf(content), false)}, nil)
				s.states.EXPECT().Get(gomock.Any(), key("web-04")).
					Return(entry{value: state("f", "/etc/a", "y", true)}, nil)
				s.states.EXPECT().Get(gomock.Any(), "bad").Return(entry{value: []byte("{")}, nil)
				s.states.EXPECT().Get(gomock.Any(), "short").Return(nil, errBoom)
				s.objects.EXPECT().GetBytes(gomock.Any(), "svc").Return(content, nil)
				s.objects.EXPECT().
					GetBytes(gomock.Any(), "ca").
					Return(nil, jetstream.ErrObjectNotFound)
				s.objects.EXPECT().GetBytes(gomock.Any(), "cron").Return(content, nil)
			},
			validateFunc: func(got file.StaleDeploymentsResponse, err error) {
				s.Require().NoError(err)
				s.Require().Equal(2, got.TotalItems)
				s.Equal(file.StaleDeployment{
					ObjectName: "svc", Hostname: "web-01", Provider: "service",
					Path: "/etc/systemd/system/a.service", DeployedSha: "old",
					CurrentSha: hexOf(content), DeployedAt: "2026-11-17T00:00:00Z",
				}, got.Items[0])
				s.Equal("certificate", got.Items[1].Provider)
				s.Empty(got.Items[1].CurrentSha)
			},
		},
		{
			name: "when a template deploy's object did not change, it is not stale",
			setup: func() {
				s.states.EXPECT().Keys(gomock.Any()).Return([]string{"web-01"}, nil)
				s.states.EXPECT().Get(gomock.Any(), "web-01").
					Return(entry{value: state("t", "/etc/app.conf", hexOf(content), false)}, nil)
				s.objects.EXPECT().GetBytes(gomock.Any(), "t").Return(content, nil)
			},
			validateFunc: func(got file.StaleDeploymentsResponse, err error) {
				s.Require().NoError(err)
				s.Empty(got.Items)
			},
		},
		{
			name: "when a template deploy's object changed, it is stale with the deployed object's checksum",
			setup: func() {
				s.states.EXPECT().Keys(gomock.Any()).Return([]string{"web-01"}, nil)
				s.states.EXPECT().Get(gomock.Any(), "web-01").
					Return(entry{value: state("t", "/etc/app.conf", "old", false)}, nil)
				s.objects.EXPECT().GetBytes(gomock.Any(), "t").Return(content, nil)
			},
			validateFunc: func(got file.StaleDeploymentsResponse, err error) {
				s.Require().NoError(err)
				s.Require().Len(got.Items, 1)
				s.Equal("old", got.Items[0].DeployedSha)
				s.Equal(hexOf(content), got.Items[0].CurrentSha)
			},
		},
		{
			name: "when a record has no object checksum, it is stale until the next deploy",
			setup: func() {
				s.states.EXPECT().Keys(gomock.Any()).Return([]string{"web-01"}, nil)
				s.states.EXPECT().Get(gomock.Any(), "web-01").Return(entry{value: []byte(
					`{"object_name":"t","path":"/etc/app.conf","sha256":"` + hexOf(content) + `"}`,
				)}, nil)
				s.objects.EXPECT().GetBytes(gomock.Any(), "t").Return(content, nil)
			},
			validateFunc: func(got file.StaleDeploymentsResponse, err error) {
				s.Require().NoError(err)
				s.Require().Len(got.Items, 1)
				s.Empty(got.Items[0].DeployedSha)
			},
		},
		{
			name: "when a record names no object, it is not stale",
			setup: func() {
				s.states.EXPECT().Keys(gomock.Any()).Return([]string{"web-01"}, nil)
				s.states.EXPECT().Get(gomock.Any(), "web-01").
					Return(entry{value: state("", "/etc/sysctl.d/osapi-a.conf", "", false)}, nil)
			},
			validateFunc: func(got file.StaleDeploymentsResponse, err error) {
				s.Require().NoError(err)
				s.Empty(got.Items)
			},
		},
		{
			name:     "when there is no state KV, it is ErrNoState",
			noStates: true,
			setup:    func() {},
			validateFunc: func(_ file.StaleDeploymentsResponse, err error) {
				s.ErrorIs(err, file.ErrNoState)
			},
		},
		{
			name:  "when the keys cannot be read, it is a 500",
			setup: func() { s.states.EXPECT().Keys(gomock.Any()).Return(nil, errBoom) },
			validateFunc: func(_ file.StaleDeploymentsResponse, err error) {
				s.EqualError(err, "failed to list file state keys")
			},
		},
		{
			name: "when a deploy is a plain file or a cron job, its provider says so",
			setup: func() {
				s.states.EXPECT().Keys(gomock.Any()).Return([]string{"web-05", "web-06"}, nil)
				s.states.EXPECT().Get(gomock.Any(), "web-05").
					Return(entry{value: state("f", "/etc/app.conf", "z", false)}, nil)
				s.states.EXPECT().Get(gomock.Any(), "web-06").
					Return(entry{value: state("c", "/etc/cron.daily/backup", "z", false)}, nil)
				s.objects.EXPECT().GetBytes(gomock.Any(), "f").Return(content, nil)
				s.objects.EXPECT().GetBytes(gomock.Any(), "c").Return(content, nil)
			},
			validateFunc: func(got file.StaleDeploymentsResponse, err error) {
				s.Require().NoError(err)
				s.Equal("file", got.Items[0].Provider)
				s.Equal("web-05", got.Items[0].Hostname)
				s.Equal("cron", got.Items[1].Provider)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			d := s.deps()
			if tc.noStates {
				d.States = nil
			}
			tc.validateFunc(file.Stale(context.Background(), d, file.StaleInput{}))
		})
	}
}

func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).Generate(signingKey, []string{"read"}, "test", perms, nil)

	return t
}

// form is a multipart body with the fields and a file.
func form(
	fields map[string]string,
	data []byte,
) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	if data != nil {
		part, _ := w.CreateFormFile("file", "upload")
		_, _ = part.Write(data)
	}
	_ = w.Close()

	return body, w.FormDataContentType()
}

// TestRoutes runs requests through the generated routes, with the auth, the
// multipart decoding, the validation and the status mapping of the runtime.
func (s *MethodsPublicTestSuite) TestRoutes() {
	tests := []struct {
		name     string
		method   string
		path     string
		fields   map[string]string
		data     []byte
		token    string
		noStore  bool
		setup    func()
		wantCode int
		wantLoc  string
	}{
		{
			name:   "when an upload is new, it is a 201 with Location",
			method: http.MethodPost,
			path:   "/api/file",
			fields: map[string]string{
				"name": "nginx.conf",
			},
			data:  content,
			token: token("file:write"),
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(nil, jetstream.ErrObjectNotFound)
				s.objects.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{}, nil)
			},
			wantCode: http.StatusCreated,
			wantLoc:  "/api/file/nginx.conf",
		},
		{
			name:     "when an upload has no file, it is a 400",
			method:   http.MethodPost,
			path:     "/api/file",
			fields:   map[string]string{"name": "nginx.conf"},
			token:    token("file:write"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "when an upload is not a form, it is a 400",
			method:   http.MethodPost,
			path:     "/api/file",
			token:    token("file:write"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "when an upload has a content type that is not raw or template, it is a 400",
			method: http.MethodPost, path: "/api/file",
			fields: map[string]string{"name": "nginx.conf", "content_type": "zip"}, data: content,
			token: token("file:write"), setup: func() {}, wantCode: http.StatusBadRequest,
		},
		{
			name:   "when an upload has other content, it is a 409",
			method: http.MethodPost,
			path:   "/api/file",
			fields: map[string]string{
				"name": "nginx.conf",
			},
			data:  content,
			token: token("file:write"),
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").
					Return(&jetstream.ObjectInfo{Digest: "SHA-256=old"}, nil)
			},
			wantCode: http.StatusConflict,
		},
		{
			name:     "when an upload names an osapi file, it is a 403",
			method:   http.MethodPost,
			path:     "/api/file",
			fields:   map[string]string{"name": "osapi/x"},
			data:     content,
			token:    token("file:write"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:   "when an update replaces a file, it is a 200",
			method: http.MethodPut,
			path:   "/api/file/nginx.conf",
			data:   content,
			token:  token("file:write"),
			setup: func() {
				s.objects.EXPECT().Info(gomock.Any(), "nginx.conf").
					Return(&jetstream.ObjectInfo{Digest: "SHA-256=old"}, nil)
				s.objects.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{}, nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name: "when the token has only file:read, an update is a 403", method: http.MethodPut,
			path: "/api/file/nginx.conf", data: content, token: token("file:read"),
			setup: func() {}, wantCode: http.StatusForbidden,
		},
		{
			name:   "when there is no Object Store, a list is a 503",
			method: http.MethodGet,
			path:   "/api/file",
			token: token(
				"file:read",
			),
			noStore:  true,
			setup:    func() {},
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name:   "when a delete succeeds, it is a 204",
			method: http.MethodDelete,
			path:   "/api/file/nginx.conf",
			token:  token("file:write"),
			setup: func() {
				s.objects.EXPECT().
					Info(gomock.Any(), "nginx.conf").
					Return(&jetstream.ObjectInfo{}, nil)
				s.objects.EXPECT().Delete(gomock.Any(), "nginx.conf").Return(nil)
			},
			wantCode: http.StatusNoContent,
		},
		{
			name:     "when there is no token, it is a 401",
			method:   http.MethodGet,
			path:     "/api/file/stale",
			setup:    func() {},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			d := s.deps()
			if tc.noStore {
				d = file.Deps{}
			}
			srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, d)
			e := echo.New()
			srv.Mount(e, gen.Routes(srv))

			var req *http.Request
			if tc.fields != nil || tc.data != nil {
				body, ct := form(tc.fields, tc.data)
				req = httptest.NewRequest(tc.method, tc.path, body)
				req.Header.Set("Content-Type", ct)
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			s.Equal(tc.wantLoc, rec.Header().Get("Location"))
		})
	}
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
