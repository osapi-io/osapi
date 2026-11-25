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
	"time"

	"github.com/osapi-io/osapi/internal/controller/jointoken"

	"github.com/avfs/avfs/vfs/osfs"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/admission"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/validation"
)

// startEnrollment loads or makes the controller keypair, and starts the
// enrollment watcher. It returns the watcher, so that the agent handler can
// list, accept and reject agents. If the enrollment bucket does not open,
// the controller does not start: PKI is always on, and an agent that is not
// admitted takes no job.
func startEnrollment(
	ctx context.Context,
	log *slog.Logger,
	conn *jobclient.Connection,
	namespace string,
	cfg config.Config,
) (*admission.Watcher, error) {
	enrollmentCfg := bus.EnrollmentKV(namespace, cfg.NATS.Enrollment)
	if err := conn.NATS.KVCreateOrUpdate(ctx, enrollmentCfg); err != nil {
		return nil, fmt.Errorf("create the enrollment KV bucket: %w", err)
	}

	keys := pki.New(osfs.New(), cfg.Controller.PKI.KeyDir, "controller")
	if err := keys.LoadOrGenerate(); err != nil {
		return nil, fmt.Errorf("load or generate controller keypair: %w", err)
	}
	// The controller verifies the job records that it signed with its own
	// key: a reader refuses a record that it cannot verify.
	keys.SetControllerPublicKey(keys.PublicKey())

	log.Info(
		"controller PKI initialized",
		slog.String("fingerprint", keys.Fingerprint()),
		slog.String("key_dir", cfg.Controller.PKI.KeyDir),
	)

	// The job client signs job payloads from now on, and verifies the
	// answers of agents as they arrive.
	conn.Jobs.SetPKISigner(keys)

	watcher := admission.NewWatcher(
		log,
		conn.NATS,
		conn.NATS,
		enrollmentCfg.Bucket,
		keys,
		cfg.Controller.PKI.AutoAdmit,
		namespace,
	)

	// Verify the answers of agents against the key recorded when each agent
	// was accepted. The watcher owns those records, so verification reads
	// what acceptance wrote and shares its cache.
	conn.Jobs.SetAgentKeyStore(watcher.KeyStore())

	// Unverified registrations stop deciding where work goes. Both sides of
	// this turn on together: the store makes verification possible.
	validation.SetEnforceVerifiedRegistrations(true)

	// A rotation replaces the key of an agent. The replaced key stays
	// acceptable for the configured window, so that a message signed just
	// before the rotation is not rejected for arriving late.
	if grace, err := time.ParseDuration(cfg.Controller.PKI.RotationGracePeriod); err == nil {
		watcher.SetRotationGracePeriod(grace)
	}

	// A join takes a use of a token in the same bucket that join token
	// create writes.
	watcher.SetJoinTokens(
		jointoken.Store{KV: conn.NATS, Bucket: enrollmentCfg.Bucket, Now: time.Now},
	)

	background(ctx, log, "enrollment watcher", watcher.Start)

	// A join over NATS: the agent sends the same signed request as over
	// HTTPS, and gets the NATS user that it connects with after the join.
	_, agentUser := cfg.JoinUsers()
	background(ctx, log, "join over NATS", func(ctx context.Context) error {
		return watcher.ServeJoin(ctx, conn.NATS.NC.QueueSubscribe, pki.JoinNATS{
			Username: agentUser.Username,
			Password: agentUser.Password,
		})
	})

	// A machine that went away without a delete keeps its key until an
	// operator removes it, unless delete_after removes it.
	if after, err := time.ParseDuration(cfg.Controller.PKI.DeleteAfter); err == nil && after > 0 {
		background(ctx, log, "stale machine cleanup", func(ctx context.Context) error {
			return watcher.RunStaleCleanup(ctx, conn.Jobs, after)
		})
	}

	return watcher, nil
}
