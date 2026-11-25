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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/natsjwt"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// NATSJoinFile is the file in agent.pki.key_dir where an agent keeps what a
// join over NATS gave it: the server, its NATS user, and the certificate of
// the server. It is 0600: it holds the user's password.
const NATSJoinFile = "nats-join.json"

// natsCreds is the file in key_dir of the agent's NATS credentials, in
// operator mode: the JWT that the join gave it, and the seed of its user,
// which never left the host. It is 0600.
const natsCreds = "nats.creds"

// natsJoinCA is the file in key_dir of the certificate that the NATS server
// showed at the join, which the agent trusts after it.
const natsJoinCA = "nats-join-ca.pem"

// errNATSJoinString is a join string that is not nats:// or tls:// with a token.
var errNATSJoinString = errors.New(
	"the join string is not nats://<host>:<port> or tls://<host>:<port> with a token",
)

// natsJoinString is a join string for NATS, parsed.
type natsJoinString struct {
	scheme      string
	host        string
	port        int
	token       string
	user        string
	password    string
	fingerprint string
	namespace   string
	// creds is, in operator mode, the bootstrap user's credentials.
	creds []byte
}

// savedJoin is what NATSJoinFile holds.
type savedJoin struct {
	Scheme    string `json:"scheme"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Namespace string `json:"namespace"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	CredsFile string `json:"creds_file,omitempty"`
	CA        string `json:"ca,omitempty"`
}

// isNATSJoin reports whether s is a join string for NATS.
func isNATSJoin(
	s string,
) bool {
	u, err := url.Parse(s)

	return err == nil && (u.Scheme == "nats" || u.Scheme == "tls")
}

// parseNATSJoin parses s:
// nats|tls://<host>:<port>?token=<t>&user=<u>&pass=<p>&fp=SHA256:<hex>&ns=<ns>.
// A tls:// string needs fp, which the agent pins.
func parseNATSJoin(
	s string,
) (natsJoinString, error) {
	// joinNATS calls it only with a string that isNATSJoin parsed.
	u, _ := url.Parse(s)
	host, portStr, err := net.SplitHostPort(u.Host)
	port, portErr := strconv.Atoi(portStr)
	q := u.Query()
	j := natsJoinString{
		scheme: u.Scheme, host: host, port: port,
		token: q.Get("token"), user: q.Get("user"), password: q.Get("pass"),
		fingerprint: q.Get("fp"), namespace: q.Get("ns"),
	}
	creds, credsErr := base64.RawURLEncoding.DecodeString(q.Get("creds"))
	j.creds = creds
	if err != nil || portErr != nil || credsErr != nil || j.token == "" ||
		(j.scheme == "tls" && j.fingerprint == "") {
		return natsJoinString{}, errNATSJoinString
	}

	return j, nil
}

// newNATSUser makes the agent's own NATS user nkey. Tests swap it.
var newNATSUser = natsjwt.NewUser

// getIdentity is the machine ID and hostname of the host. Tests swap it.
var getIdentity = identity.GetIdentity

// natsRequest connects to url with opts, sends data to subject as a
// request, and returns the reply. Tests swap it.
var natsRequest = func(
	url string,
	opts []nats.Option,
	subject string,
	data []byte,
) ([]byte, error) {
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}
	defer nc.Close()
	msg, err := nc.Request(subject, data, 15*time.Second)
	if err != nil {
		return nil, err
	}

	return msg.Data, nil
}

// joinNATS gives cfg the NATS settings of a join over NATS, when
// agent.join is a join string for NATS: the ones that an earlier join saved
// in key_dir, or else the ones that a join now gives. The agent then
// connects with its own user, and agent.join is spent. A config whose
// agent.join is not for NATS is cfg.
func joinNATS(
	ctx context.Context,
	log *slog.Logger,
	fs avfs.VFS,
	cfg config.Config,
) (config.Config, error) {
	if !isNATSJoin(cfg.Agent.Join) {
		return cfg, nil
	}
	keyDir := cfg.Agent.PKI.KeyDir
	saved, err := readJoin(fs, keyDir)
	if err != nil {
		// The span names the server, never the token or the users.
		spanCtx, span := opspan.Start(ctx, "osapi-agent", "agent.join",
			attribute.String("osapi.join.transport", "nats"),
		)
		saved, err = joinNow(spanCtx, log, fs, cfg)
		if err := opspan.End(span, err); err != nil {
			return cfg, fmt.Errorf("join over NATS: %w", err)
		}
	}

	return applyJoin(fs, cfg, saved)
}

// readJoin is the join that key_dir keeps.
func readJoin(
	fs avfs.VFS,
	keyDir string,
) (savedJoin, error) {
	data, err := avfs.ReadFile(fs, filepath.Join(keyDir, NATSJoinFile))
	if err != nil {
		return savedJoin{}, err
	}
	var s savedJoin
	if err := json.Unmarshal(data, &s); err != nil {
		return savedJoin{}, fmt.Errorf("%s: %w", NATSJoinFile, err)
	}

	return s, nil
}

// joinNow sends the signed join request with the bootstrap user of the join
// string, and keeps what the controller replies: its key, as the agent's
// admission, and the agent's NATS user, in NATSJoinFile.
func joinNow(
	ctx context.Context,
	log *slog.Logger,
	fs avfs.VFS,
	cfg config.Config,
) (savedJoin, error) {
	j, err := parseNATSJoin(cfg.Agent.Join)
	if err != nil {
		return savedJoin{}, err
	}
	if j.namespace == "" {
		j.namespace = cfg.Agent.NATS.Namespace
	}
	id, err := getIdentity(fs, cfg.Agent.Hostname)
	if err != nil {
		return savedJoin{}, err
	}
	keyDir := cfg.Agent.PKI.KeyDir
	keys := pki.New(fs, keyDir, "agent", pki.WithEncryptionKey())
	if err := keys.LoadOrGenerate(); err != nil {
		return savedJoin{}, fmt.Errorf("load or make the agent's keys: %w", err)
	}
	// The agent's own NATS user, for operator mode: the controller mints a
	// JWT for its public key, and the seed stays here.
	userSeed, userKey, err := newNATSUser()
	if err != nil {
		return savedJoin{}, err
	}
	req := pki.JoinRequest{
		Token:         j.token,
		MachineID:     id.MachineID,
		Hostname:      id.Hostname,
		PublicKey:     keys.PublicKey(),
		EncryptionKey: keys.EncryptionPublicKey(),
		NATSUserKey:   userKey,
	}
	req.Sign(keys.PrivateKey())
	data, _ := json.Marshal(req)

	var seen []*x509.Certificate
	opts := []nats.Option{nats.Name("osapi-agent-join"), nats.Timeout(10 * time.Second)}
	if j.user != "" {
		opts = append(opts, nats.UserInfo(j.user, j.password))
	}
	if len(j.creds) > 0 {
		opts = append(opts, nats.UserCredentialBytes(j.creds))
	}
	if j.scheme == "tls" {
		opts = append(opts, nats.Secure(&tls.Config{
			MinVersion: tls.VersionTLS12,
			// The fingerprint of the join string, which VerifyConnection
			// checks, takes the place of the roots.
			InsecureSkipVerify: true, //nolint:gosec // VerifyConnection pins the certificate
			VerifyConnection: func(cs tls.ConnectionState) error {
				seen = cs.PeerCertificates

				return verifyPin(j.fingerprint, cs.PeerCertificates)
			},
		}))
	}
	subject := pki.JoinRequestSuffix
	if j.namespace != "" {
		subject = j.namespace + "." + subject
	}
	addr := fmt.Sprintf("%s://%s", j.scheme, net.JoinHostPort(j.host, strconv.Itoa(j.port)))
	log.Info(
		"joining over NATS",
		slog.String("server", addr),
		slog.String("machine_id", id.MachineID),
	)
	raw, err := natsRequest(addr, opts, subject, data)
	if err != nil {
		return savedJoin{}, fmt.Errorf("%s: %w", addr, err)
	}
	var reply pki.JoinReply
	if err := json.Unmarshal(raw, &reply); err != nil {
		return savedJoin{}, fmt.Errorf("the reply of the join: %w", err)
	}
	if reply.Error != "" {
		return savedJoin{}, errors.New(reply.Error)
	}

	keyPEM := pem.EncodeToMemory(
		&pem.Block{Type: "ED25519 PUBLIC KEY", Bytes: reply.ControllerPublicKey},
	)
	if err := atomicfile.Write(
		ctx,
		fs,
		filepath.Join(keyDir, EnrolledControllerKey),
		keyPEM,
		0o644,
	); err != nil {
		return savedJoin{}, err
	}
	s := savedJoin{
		Scheme: j.scheme, Host: j.host, Port: j.port, Namespace: j.namespace,
		Username: reply.NATS.Username, Password: reply.NATS.Password,
	}
	if reply.NATS.JWT != "" {
		creds, err := natsjwt.Creds(reply.NATS.JWT, userSeed)
		if err != nil {
			return savedJoin{}, fmt.Errorf("the JWT of the join: %w", err)
		}
		s.CredsFile = filepath.Join(keyDir, natsCreds)
		if err := atomicfile.Write(ctx, fs, s.CredsFile, creds, 0o600); err != nil {
			return savedJoin{}, err
		}
	}
	if len(seen) > 0 {
		s.CA = string(
			pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: seen[len(seen)-1].Raw}),
		)
	}
	out, _ := json.Marshal(s)
	if err := atomicfile.Write(
		ctx,
		fs,
		filepath.Join(keyDir, NATSJoinFile),
		out,
		0o600,
	); err != nil {
		return savedJoin{}, err
	}
	log.Info("joined over NATS, agent admitted", slog.String("machine_id", id.MachineID))

	return s, nil
}

// applyJoin is cfg with the NATS settings of s.
func applyJoin(
	fs avfs.VFS,
	cfg config.Config,
	s savedJoin,
) (config.Config, error) {
	n := &cfg.Agent.NATS
	n.Host, n.Port, n.Namespace = s.Host, s.Port, s.Namespace
	n.Auth = config.NATSAuth{Type: "none"}
	switch {
	case s.CredsFile != "":
		n.Auth = config.NATSAuth{Type: "creds", CredsFile: s.CredsFile}
	case s.Username != "":
		n.Auth = config.NATSAuth{Type: "user_pass", Username: s.Username, Password: s.Password}
	}
	n.TLS = config.ClientTLS{}
	if s.Scheme == "tls" {
		n.TLS.Enabled = true
		if s.CA != "" {
			ca := filepath.Join(cfg.Agent.PKI.KeyDir, natsJoinCA)
			if err := atomicfile.Write(
				context.Background(),
				fs,
				ca,
				[]byte(s.CA),
				0o644,
			); err != nil {
				return cfg, err
			}
			n.TLS.CAFile = ca
		}
	}
	cfg.Agent.Join = ""

	return cfg, nil
}
