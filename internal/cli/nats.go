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

// natsCloser is implemented by any NATS client that can be closed.
type natsCloser interface {
	Close()
}

// CloseNATSClient safely closes a NATS client connection.
func CloseNATSClient(
	nc natsCloser,
) {
	nc.Close()
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

// BuildJobKVConfig builds a jetstream.KeyValueConfig from job KV config values.
// The returned config includes TTL, MaxBytes, Storage, and Replicas from the
// NATSKV configuration.
func BuildJobKVConfig(
	namespace string,
	kvCfg config.NATSKV,
) jetstream.KeyValueConfig {
	bucket := job.ApplyNamespaceToInfraName(namespace, kvCfg.Bucket)
	ttl, _ := time.ParseDuration(kvCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   bucket,
		TTL:      ttl,
		MaxBytes: kvCfg.MaxBytes,
		Storage:  ParseJetstreamStorageType(kvCfg.Storage),
		Replicas: kvCfg.Replicas,
	}
}

// BuildResponseKVConfig builds a jetstream.KeyValueConfig for the job response
// KV bucket. It shares TTL, MaxBytes, Storage, and Replicas settings from the
// parent NATSKV configuration.
func BuildResponseKVConfig(
	namespace string,
	kvCfg config.NATSKV,
) jetstream.KeyValueConfig {
	bucket := job.ApplyNamespaceToInfraName(namespace, kvCfg.ResponseBucket)
	ttl, _ := time.ParseDuration(kvCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   bucket,
		TTL:      ttl,
		MaxBytes: kvCfg.MaxBytes,
		Storage:  ParseJetstreamStorageType(kvCfg.Storage),
		Replicas: kvCfg.Replicas,
	}
}

// BuildRegistryKVConfig builds a jetstream.KeyValueConfig from registry config values.
func BuildRegistryKVConfig(
	namespace string,
	registryCfg config.NATSRegistry,
) jetstream.KeyValueConfig {
	registryBucket := job.ApplyNamespaceToInfraName(namespace, registryCfg.Bucket)
	registryTTL, _ := time.ParseDuration(registryCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   registryBucket,
		TTL:      registryTTL,
		Storage:  ParseJetstreamStorageType(registryCfg.Storage),
		Replicas: registryCfg.Replicas,
	}
}

// BuildFactsKVConfig builds a jetstream.KeyValueConfig from facts config values.
func BuildFactsKVConfig(
	namespace string,
	factsCfg config.NATSFacts,
) jetstream.KeyValueConfig {
	factsBucket := job.ApplyNamespaceToInfraName(namespace, factsCfg.Bucket)
	factsTTL, _ := time.ParseDuration(factsCfg.TTL)

	return jetstream.KeyValueConfig{
		Bucket:   factsBucket,
		TTL:      factsTTL,
		Storage:  ParseJetstreamStorageType(factsCfg.Storage),
		Replicas: factsCfg.Replicas,
	}
}

// BuildStateKVConfig builds a jetstream.KeyValueConfig from state config values.
// The state bucket has no TTL so drain flags and timeline events persist indefinitely.
func BuildStateKVConfig(
	namespace string,
	stateCfg config.NATSState,
) jetstream.KeyValueConfig {
	stateBucket := job.ApplyNamespaceToInfraName(namespace, stateCfg.Bucket)

	return jetstream.KeyValueConfig{
		Bucket:   stateBucket,
		Storage:  ParseJetstreamStorageType(stateCfg.Storage),
		Replicas: stateCfg.Replicas,
	}
}

// BuildAuditStreamConfig builds a jetstream.StreamConfig from audit
// config values.
func BuildAuditStreamConfig(
	namespace string,
	auditCfg config.NATSAudit,
) jetstream.StreamConfig {
	streamName := job.ApplyNamespaceToInfraName(
		namespace,
		auditCfg.Stream,
	)
	subject := job.ApplyNamespaceToSubjects(
		namespace,
		auditCfg.Subject,
	)
	maxAge, _ := time.ParseDuration(auditCfg.MaxAge)

	return jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject + ".>"},
		MaxAge:   maxAge,
		MaxBytes: auditCfg.MaxBytes,
		Storage:  ParseJetstreamStorageType(auditCfg.Storage),
		Replicas: auditCfg.Replicas,
		Discard:  jetstream.DiscardOld,
	}
}

// BuildObjectStoreConfig builds a jetstream.ObjectStoreConfig from objects config values.
func BuildObjectStoreConfig(
	namespace string,
	objectsCfg config.NATSObjects,
) jetstream.ObjectStoreConfig {
	objectsBucket := job.ApplyNamespaceToInfraName(namespace, objectsCfg.Bucket)

	return jetstream.ObjectStoreConfig{
		Bucket:   objectsBucket,
		MaxBytes: objectsCfg.MaxBytes,
		Storage:  ParseJetstreamStorageType(objectsCfg.Storage),
		Replicas: objectsCfg.Replicas,
	}
}

// BuildFileStateKVConfig builds a jetstream.KeyValueConfig from file-state config values.
// The file-state bucket has no TTL so deployment SHA tracking persists indefinitely.
func BuildFileStateKVConfig(
	namespace string,
	fileStateCfg config.NATSFileState,
) jetstream.KeyValueConfig {
	fileStateBucket := job.ApplyNamespaceToInfraName(namespace, fileStateCfg.Bucket)

	return jetstream.KeyValueConfig{
		Bucket:   fileStateBucket,
		Storage:  ParseJetstreamStorageType(fileStateCfg.Storage),
		Replicas: fileStateCfg.Replicas,
	}
}

// BuildEnrollmentKVConfig builds a jetstream.KeyValueConfig from enrollment config values.
// The enrollment bucket has no TTL so pending requests persist until accepted or rejected.
func BuildEnrollmentKVConfig(
	namespace string,
	enrollmentCfg config.NATSEnrollment,
) jetstream.KeyValueConfig {
	enrollmentBucket := job.ApplyNamespaceToInfraName(namespace, enrollmentCfg.Bucket)

	return jetstream.KeyValueConfig{
		Bucket:   enrollmentBucket,
		Storage:  ParseJetstreamStorageType(enrollmentCfg.Storage),
		Replicas: enrollmentCfg.Replicas,
	}
}
