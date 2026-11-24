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

package controller

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// openVaults makes each vault of controller.secrets.vaults from its type.
// A vault that does not open, such as one with a type that osapi does not
// know or a key that is not there, stops the controller, because a secret
// that it cannot resolve would fail each request that uses it.
func openVaults(
	ctx context.Context,
	log *slog.Logger,
	conn *jobclient.Connection,
	cfg config.Config,
) (*vault.Set, error) {
	if err := makeVaultKeys(
		ctx,
		log,
		osfs.NewWithNoIdm(),
		cfg.Controller.Secrets.Vaults,
	); err != nil {
		return nil, err
	}
	set, err := OpenVaults(ctx, conn.NATS, cfg)
	if err != nil {
		return nil, err
	}
	for _, name := range set.Names() {
		spec, _ := set.Spec(name)
		log.Info("vault opened",
			slog.String("name", name),
			slog.String("type", spec.Type),
			slog.Any("targets", spec.Targets))
	}

	return set, nil
}

// OpenVaults makes each vault of controller.secrets.vaults, with nc as
// the NATS connection of a type that needs one. A command that resolves a
// reference of the config, such as osapi controller token generate, opens
// the vaults with it, as the controller does.
func OpenVaults(
	ctx context.Context,
	nc *natsclient.Client,
	cfg config.Config,
) (*vault.Set, error) {
	var bucket string
	if cfg.NATS.Secrets.Bucket != "" {
		bucket = job.ApplyNamespaceToInfraName(
			cfg.Controller.NATS.Namespace,
			cfg.NATS.Secrets.Bucket,
		)
	}

	return vault.Open(ctx, vault.Env{
		NATS:      nc,
		Bucket:    bucket,
		FS:        osfs.NewWithNoIdm(),
		LookupEnv: os.LookupEnv,
	}, cfg.Controller.Secrets.Vaults)
}

// makeVaultKeys makes the key file of each nats vault whose key_file is not
// there, as osapi controller vault key create does, so the default vault
// works at the first start. It never replaces a key that is there: each
// value under it would then not decrypt. A vault whose key is in key_env
// gets no file.
func makeVaultKeys(
	ctx context.Context,
	log *slog.Logger,
	vfs avfs.VFS,
	specs []vault.Spec,
) error {
	for _, spec := range specs {
		if spec.Type != natsvault.TypeName || spec.KeyFile == "" || spec.KeyEnv != "" {
			continue
		}
		if _, err := vfs.Stat(spec.KeyFile); err == nil {
			continue
		}
		ring, err := natsvault.CreateKey(ctx, vfs, spec.KeyFile)
		if err != nil {
			return fmt.Errorf("vault %q: make its key: %w", spec.Name, err)
		}
		log.Info("vault key created",
			slog.String("name", spec.Name),
			slog.String("key_file", spec.KeyFile),
			slog.String("key_id", ring.ID()))
	}

	return nil
}
