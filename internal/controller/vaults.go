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
	"log/slog"
	"os"

	"github.com/avfs/avfs/vfs/osfs"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/secret/vault"

	// The nats type of vault, which controller.secrets.vaults can name.
	_ "github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// openVaults makes each vault of controller.secrets.vaults from its type.
// A vault that does not open, such as one with a type that osapi does not
// know or a key that is not there, stops the controller, because a secret
// that it cannot resolve would fail each request that uses it.
func openVaults(
	ctx context.Context,
	log *slog.Logger,
	conn *jobclient.Connection,
	namespace string,
	cfg config.Config,
) (*vault.Set, error) {
	var bucket string
	if cfg.NATS.Secrets.Bucket != "" {
		bucket = job.ApplyNamespaceToInfraName(namespace, cfg.NATS.Secrets.Bucket)
	}
	set, err := vault.Open(ctx, vault.Env{
		NATS:      conn.NATS,
		Bucket:    bucket,
		FS:        osfs.NewWithNoIdm(),
		LookupEnv: os.LookupEnv,
	}, cfg.Controller.Secrets.Vaults)
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
