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

// Package cli provides shared utilities for CLI startup commands.
package cli

import (
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/messaging"
)

// ParseJetstreamStorageType maps "memory"/"file" strings to jetstream.StorageType.
func ParseJetstreamStorageType(
	s string,
) jetstream.StorageType {
	if s == "memory" {
		return jetstream.MemoryStorage
	}

	return jetstream.FileStorage
}

// CloseNATSClient safely closes a NATS client connection.
func CloseNATSClient(
	nc messaging.NATSClient,
) {
	if natsConn, ok := nc.(*natsclient.Client); ok && natsConn.NC != nil {
		natsConn.NC.Close()
	}
}

// BuildNATSAuthOptions converts a config NATSAuth to natsclient.AuthOptions.
func BuildNATSAuthOptions(
	auth config.NATSAuth,
) natsclient.AuthOptions {
	switch auth.Type {
	case "user_pass":
		return natsclient.AuthOptions{
			AuthType: natsclient.UserPassAuth,
			Username: auth.Username,
			Password: auth.Password,
		}
	case "nkey":
		return natsclient.AuthOptions{
			AuthType: natsclient.NKeyAuth,
			NKeyFile: auth.NKeyFile,
		}
	default:
		return natsclient.AuthOptions{
			AuthType: natsclient.NoAuth,
		}
	}
}

// BuildAuditKVConfig builds a jetstream.KeyValueConfig from audit config values.
func BuildAuditKVConfig(
	namespace string,
	auditCfg config.NATSAudit,
) jetstream.KeyValueConfig {
	auditBucket := job.ApplyNamespaceToInfraName(namespace, auditCfg.Bucket)
	auditTTL, _ := time.ParseDuration(auditCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   auditBucket,
		TTL:      auditTTL,
		MaxBytes: auditCfg.MaxBytes,
		Storage:  ParseJetstreamStorageType(auditCfg.Storage),
		Replicas: auditCfg.Replicas,
	}
}
