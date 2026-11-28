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
	"net"
	"strconv"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
)

// ErrRotateRefused is a rotation that the controller refused. The new keys
// are removed, and the agent keeps its keys.
var ErrRotateRefused = errors.New("the controller refused the rotation")

// RotateKey replaces the agent's two keys with new ones that the controller
// takes in place of the admitted ones, with no new admission, and returns
// the fingerprint of the new keys.
//
// It makes the new keys beside the old ones, in agent.next.*, and sends the
// new public keys to the controller over NATS, as the agent connects,
// signed with the old signing key and with the new one. When the controller
// takes them, it moves the new keys in place of the old ones; a running
// agent reads them at its next heartbeat. When the controller refuses, it
// removes the new keys. When the controller does not answer, it keeps them:
// the controller may have taken them, and a second rotation sends the same
// keys again.
func RotateKey(
	ctx context.Context,
	log *slog.Logger,
	fs avfs.VFS,
	cfg config.Config,
) (string, error) {
	keyDir := cfg.Agent.PKI.KeyDir
	keys := pki.New(fs, keyDir, "agent", pki.WithEncryptionKey())
	if err := loadAgentKeys(ctx, log, keys, keys.Load); err != nil {
		return "", fmt.Errorf("the agent's keys: %w; the agent makes them at its first start", err)
	}
	next, err := keys.Next()
	if err != nil {
		return "", fmt.Errorf("make the new keys: %w", err)
	}
	id, err := getIdentity(fs, cfg.Agent.Hostname)
	if err != nil {
		return "", fmt.Errorf("the identity of this machine: %w", err)
	}
	req := pki.RotateRequest{
		MachineID:     id.MachineID,
		PublicKey:     next.PublicKey(),
		EncryptionKey: next.EncryptionPublicKey(),
		At:            renewNow().Unix(),
	}
	req.Sign(keys.PrivateKey(), next.PrivateKey())
	data, _ := json.Marshal(req)

	if s, err := readJoin(fs, keyDir); err == nil {
		if cfg, err = applyJoin(ctx, fs, cfg, s); err != nil {
			return "", err
		}
	}
	addr, opts, err := natsDialOptions("osapi-agent-rotate", cfg.Agent.NATS)
	if err != nil {
		return "", err
	}
	subject := pki.RotateRequestSuffix
	if ns := cfg.Agent.NATS.Namespace; ns != "" {
		subject = ns + "." + subject
	}
	raw, err := natsRequest(ctx, addr, opts, subject, data)
	if err != nil {
		return "", fmt.Errorf(
			"%s: %w; the new keys stay in %s, and a second rotation sends them again",
			addr, err, keyDir,
		)
	}
	var reply pki.RotateReply
	if err := json.Unmarshal(raw, &reply); err != nil {
		return "", fmt.Errorf("the reply of the rotation: %w", err)
	}
	if reply.Error != "" {
		if err := next.Drop(); err != nil {
			log.WarnContext(
				ctx,
				"failed to remove the refused keys",
				slog.String("error", err.Error()),
			)
		}

		return "", fmt.Errorf("%w: %s", ErrRotateRefused, reply.Error)
	}
	if err := keys.Promote(next); err != nil {
		return "", fmt.Errorf(
			"%w; the controller took the new keys, so run the rotation again to move them in place",
			err,
		)
	}
	log.InfoContext(ctx, "rotated the agent's keys", slog.String("fingerprint", reply.Fingerprint))

	return keys.Fingerprint(), nil
}

// loadAgentKeys finishes a key rotation that stopped between its renames,
// then loads the agent's keys with load, such as Load or LoadOrGenerate.
// Without the repair, a stop between the renames of Promote leaves the
// private key of one pair beside the public key of the other, and the agent
// signs with a key that the controller does not verify.
func loadAgentKeys(
	ctx context.Context,
	log *slog.Logger,
	keys *pki.Manager,
	load func() error,
) error {
	moved, err := keys.ResumeNext()
	if err != nil {
		return fmt.Errorf("finish the rotation of the agent's keys: %w", err)
	}
	if moved {
		log.InfoContext(
			ctx,
			"finished a rotation of the agent's keys that stopped between its renames",
		)
	}

	return load()
}

// natsDialOptions is the URL and the options of a connection named name
// to NATS from conn, as nats-client makes them.
func natsDialOptions(
	name string,
	conn config.NATSConnection,
) (string, []nats.Option, error) {
	tlsConfig, err := conn.TLS.Dial(conn.Host)
	if err != nil {
		return "", nil, fmt.Errorf("nats tls: %w", err)
	}
	scheme := "nats"
	opts := []nats.Option{nats.Name(name), nats.Timeout(10 * time.Second)}
	if tlsConfig != nil {
		scheme = "tls"
		opts = append(opts, nats.Secure(tlsConfig))
	}
	switch conn.Auth.Type {
	case "user_pass":
		opts = append(opts, nats.UserInfo(conn.Auth.Username, conn.Auth.Password))
	case "nkey":
		opt, err := nats.NkeyOptionFromSeed(conn.Auth.NKeyFile)
		if err != nil {
			return "", nil, fmt.Errorf("nats nkey: %w", err)
		}
		opts = append(opts, opt)
	case "creds":
		opts = append(opts, nats.UserCredentials(conn.Auth.CredsFile))
	}

	return fmt.Sprintf(
		"%s://%s",
		scheme,
		net.JoinHostPort(conn.Host, strconv.Itoa(conn.Port)),
	), opts, nil
}
