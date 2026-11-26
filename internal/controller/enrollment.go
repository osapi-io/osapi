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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nkeys"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/osapi-io/osapi/internal/natscallout"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/osapi-io/osapi/internal/controller/jointoken"

	"github.com/avfs/avfs/vfs/osfs"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/admission"
	jobclient "github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/internal/natsjwt"
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
	auditLog audit.Writer,
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
	if auditLog != nil {
		conn.Jobs.SetAudit(auditLog)
	}

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
	tokens := jointoken.Store{KV: conn.NATS, Bucket: enrollmentCfg.Bucket, Now: time.Now}
	watcher.SetJoinTokens(tokens)

	// Each admission decision is an event of the machine's timeline.
	watcher.SetTimeline(conn.Jobs)

	background(ctx, log, "enrollment watcher", watcher.Start)

	// A join over NATS: the agent sends the same signed request as over
	// HTTPS, and gets the NATS user that it connects with after the join.
	if auditLog != nil {
		watcher.SetAudit(auditLog)
	}
	agentUser, mint, err := joinAgentUser(cfg, namespace)
	if err != nil {
		return nil, fmt.Errorf("join over NATS: %w", err)
	}
	background(ctx, log, "join over NATS", func(ctx context.Context) error {
		return watcher.ServeJoin(ctx, conn.NATS.NC.QueueSubscribe, agentUser)
	})
	// With a NATS server's auth callout, the controller decides each
	// connection of a join or an agent.
	if cfg.Controller.Join.Callout.IssuerKeyFile != "" {
		svc, err := calloutService(cfg, namespace, watcher, tokens, log)
		if err != nil {
			return nil, fmt.Errorf("auth callout: %w", err)
		}
		background(ctx, log, "auth callout", func(ctx context.Context) error {
			c, err := calloutConnect(log, cfg)
			if err != nil {
				return err
			}
			defer bus.Close(c.NC)

			return svc.Serve(ctx, c.NC.Subscribe)
		})
	}
	// In operator mode an agent renews the JWT of its NATS user; only an
	// admitted machine gets a new one.
	if mint != nil {
		background(ctx, log, "renewal over NATS", func(ctx context.Context) error {
			return watcher.ServeRenew(ctx, conn.NATS.NC.QueueSubscribe, mint)
		})
	}

	// An admitted agent rotates its keys over NATS, in each auth mode.
	background(ctx, log, "key rotation over NATS", func(ctx context.Context) error {
		return watcher.ServeRotate(ctx, conn.NATS.NC.QueueSubscribe)
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

// errNoUserKey is a join in operator mode from an agent that sent no NATS
// user key: an agent older than the controller.
var errNoUserKey = errors.New("the agent sent no NATS user key: upgrade it")

// joinAgentUser is the NATS user that a join gives an agent: in operator
// mode, a JWT for the agent's own user key, with the agent's permissions,
// signed with controller.join.signing_key_file; else the static agent user.
func joinAgentUser(
	cfg config.Config,
	namespace string,
) (admission.AgentUser, admission.Mint, error) {
	j := cfg.Controller.Join
	// With auth callout, the agent connects with its own user key, which
	// the callout admits as the machine: the join gives it no user.
	if j.Callout.IssuerKeyFile != "" {
		return func(req pki.JoinRequest) (pki.JoinNATS, error) {
			if req.NATSUserKey == "" {
				return pki.JoinNATS{}, errNoUserKey
			}

			return pki.JoinNATS{Callout: true}, nil
		}, nil, nil
	}
	if j.SigningKeyFile == "" {
		_, user := cfg.JoinUsers()

		return admission.StaticUser(
			pki.JoinNATS{Username: user.Username, Password: user.Password},
		), nil, nil
	}
	signer, err := natsjwt.Load(j.SigningKeyFile, j.Account)
	if err != nil {
		return nil, nil, err
	}
	// A scoped key's template sets the permissions, and a JWT that it signs
	// carries none.
	var perms bus.Permissions
	if !j.Scoped {
		perms, _ = bus.RolePermissions(bus.RoleAgent, namespace, cfg.NATS)
	}
	// The config's check refuses an agent_ttl that is not a duration; an
	// empty one is no expiry.
	ttl, _ := time.ParseDuration(j.AgentTTL)
	mint := func(userKey string, machineID string) (string, error) {
		var expires time.Time
		if ttl > 0 {
			expires = time.Now().Add(ttl)
		}

		return signer.Mint(userKey, machineID, perms, expires)
	}

	return func(req pki.JoinRequest) (pki.JoinNATS, error) {
		if req.NATSUserKey == "" {
			return pki.JoinNATS{}, errNoUserKey
		}
		token, err := mint(req.NATSUserKey, req.MachineID)
		if err != nil {
			return pki.JoinNATS{}, err
		}

		return pki.JoinNATS{JWT: token}, nil
	}, mint, nil
}

// calloutService is osapi's auth callout: the issuer key of
// controller.join.callout, and the decisions of the watcher, with the join
// and agent roles of the config.
func calloutService(
	cfg config.Config,
	namespace string,
	watcher *admission.Watcher,
	tokens admission.TokenChecker,
	log *slog.Logger,
) (natscallout.Service, error) {
	c := cfg.Controller.Join.Callout
	data, err := os.ReadFile(c.IssuerKeyFile)
	if err != nil {
		return natscallout.Service{}, fmt.Errorf("read the issuer key: %w", err)
	}
	issuer, err := nkeys.FromSeed([]byte(strings.TrimSpace(string(data))))
	if err != nil {
		return natscallout.Service{}, fmt.Errorf("the issuer key: %w", err)
	}
	if pub, _ := issuer.PublicKey(); !nkeys.IsValidPublicAccountKey(pub) {
		return natscallout.Service{}, errors.New("the issuer key is not an account key (SA...)")
	}
	join, _ := bus.RolePermissions(bus.RoleJoin, namespace, cfg.NATS)
	agent, _ := bus.RolePermissions(bus.RoleAgent, namespace, cfg.NATS)
	// The config's check refuses an agent_ttl that is not a duration.
	ttl, _ := time.ParseDuration(cfg.Controller.Join.AgentTTL)

	return natscallout.Service{
		Issuer:  issuer,
		Account: c.Account,
		Logger:  log,
		Decide: watcher.Callout(admission.CalloutRoles{
			Join: join, Agent: agent, AgentTTL: ttl, Tokens: tokens,
		}),
	}, nil
}

// calloutConnect connects to the controller's NATS as the callout user,
// one of the server's auth_users, which the callout does not decide.
var calloutConnect = func(
	log *slog.Logger,
	cfg config.Config,
) (*natsclient.Client, error) {
	n := cfg.Controller.NATS
	n.Auth = cfg.Controller.Join.Callout.Auth
	opts, err := bus.ConnectionOptions(n, "osapi-auth-callout")
	if err != nil {
		return nil, err
	}

	return bus.Dial(log, opts)
}
