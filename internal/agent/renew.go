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

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"

	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/natsjwt"
	"github.com/osapi-io/osapi/internal/pki"
)

// renewCheck is how often the agent looks at the expiry of its JWT. Tests
// swap it.
var renewCheck = time.Minute

// renewNow is the time. Tests swap it.
var renewNow = time.Now

// writeCreds writes the agent's renewed credentials. Tests swap it.
var writeCreds = atomicfile.Write

// startRenewal renews the JWT of the agent's NATS user, in operator mode,
// until the agent stops: when a third of its life is left, the agent asks
// the controller for a new one. An agent that joined with static users, or
// whose JWT never expires, renews nothing.
func (a *Agent) startRenewal(
	ctx context.Context,
) {
	s, err := readJoin(a.appFs, a.appConfig.Agent.PKI.KeyDir)
	if err != nil || s.CredsFile == "" {
		return
	}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(renewCheck)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := a.renewIfDue(s); err != nil {
					a.logger.Warn(
						"failed to renew the NATS user",
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}()
}

// renewIfDue renews the JWT of s.CredsFile when a third of its life is
// left.
func (a *Agent) renewIfDue(
	s savedJoin,
) error {
	creds, err := avfs.ReadFile(a.appFs, s.CredsFile)
	if err != nil {
		return fmt.Errorf("read the NATS creds: %w", err)
	}
	// ParseDecoratedJWT gives text that is not a creds file as it is: the
	// decode below refuses it.
	userJWT, _ := jwt.ParseDecoratedJWT(creds)
	uc, err := jwt.DecodeUserClaims(userJWT)
	if err != nil {
		return fmt.Errorf("%s: %w", s.CredsFile, err)
	}
	now := renewNow().Unix()
	if uc.Expires == 0 || uc.Expires-now > (uc.Expires-uc.IssuedAt)/3 {
		return nil
	}
	kp, err := jwt.ParseDecoratedNKey(creds)
	if err != nil {
		return fmt.Errorf("%s: %w", s.CredsFile, err)
	}
	seed, _ := kp.Seed()
	userKey, _ := kp.PublicKey()

	a.reloadKeys()
	req := pki.RenewRequest{MachineID: a.machineID, NATSUserKey: userKey, At: now}
	req.Sign(a.pkiManager.PrivateKey())
	data, _ := json.Marshal(req)
	cfg, err := applyJoin(a.appFs, a.appConfig, s)
	if err != nil {
		return fmt.Errorf("the NATS settings of the join: %w", err)
	}
	// The renewal connects with the creds that it read, not the file.
	cfg.Agent.NATS.Auth = config.NATSAuth{Type: "none"}
	addr, opts, err := natsDialOptions("osapi-agent-renew", cfg.Agent.NATS)
	if err != nil {
		return fmt.Errorf("the NATS connection of the renewal: %w", err)
	}
	opts = append(opts, nats.UserCredentialBytes(creds))
	subject := pki.RenewRequestSuffix
	if s.Namespace != "" {
		subject = s.Namespace + "." + subject
	}
	raw, err := natsRequest(addr, opts, subject, data)
	if err != nil {
		return fmt.Errorf("%s: %w", addr, err)
	}
	var reply pki.JoinReply
	if err := json.Unmarshal(raw, &reply); err != nil {
		return fmt.Errorf("the reply of the renewal: %w", err)
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	nc, err := jwt.DecodeUserClaims(reply.NATS.JWT)
	if err != nil {
		return fmt.Errorf("the JWT of the renewal: %w", err)
	}
	if nc.Expires != 0 && nc.Expires <= now {
		return errors.New("the JWT of the renewal has expired")
	}
	next, err := natsjwt.Creds(reply.NATS.JWT, seed)
	if err != nil {
		return fmt.Errorf("the JWT of the renewal: %w", err)
	}
	if err := writeCreds(
		context.Background(),
		a.appFs,
		s.CredsFile,
		next,
		0o600,
	); err != nil {
		return fmt.Errorf("write the renewed NATS creds: %w", err)
	}
	a.logger.Info("renewed the NATS user")

	return nil
}
