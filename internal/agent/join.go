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
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// errJoinString is a join string that is not osapi://host:port?token=...&fp=SHA256:hex.
var errJoinString = errors.New(
	"the join string is not osapi://<host>:<port> with a token and a SHA256 fingerprint",
)

// errFingerprint is a controller whose certificate is not the one that the
// join string pins.
var errFingerprint = errors.New(
	"the controller's certificate does not match the join string's fingerprint",
)

// joinString is a join string, parsed: the controller's address, the token,
// and the fingerprint of its certificate.
type joinString struct {
	host        string
	token       string
	fingerprint string
}

// parseJoin parses s, osapi://<host>:<port>?token=<token>&fp=SHA256:<hex>.
func parseJoin(
	s string,
) (joinString, error) {
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "osapi" || u.Host == "" {
		return joinString{}, errJoinString
	}
	j := joinString{host: u.Host, token: u.Query().Get("token"), fingerprint: u.Query().Get("fp")}
	if j.token == "" || !strings.HasPrefix(j.fingerprint, "SHA256:") {
		return joinString{}, errJoinString
	}

	return j, nil
}

// pinned is a transport that connects only to the server whose certificate
// has fingerprint: the join string names the controller, so the agent checks
// it before it sends the token, and a host in the path cannot answer for it.
func pinned(
	fingerprint string,
) http.RoundTripper {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		// The fingerprint, which VerifyConnection checks, takes the place of
		// the roots: the controller's certificate can be self-signed.
		InsecureSkipVerify: true, //nolint:gosec // VerifyConnection pins the certificate
		VerifyConnection: func(cs tls.ConnectionState) error {
			return verifyPin(fingerprint, cs.PeerCertificates)
		},
	}

	return t
}

// verifyPin checks that the server's certificate, the first of certs, has
// fingerprint.
func verifyPin(
	fingerprint string,
	certs []*x509.Certificate,
) error {
	if len(certs) == 0 {
		return errFingerprint
	}
	sum := sha256.Sum256(certs[0].Raw)
	if "SHA256:"+hex.EncodeToString(sum[:]) != fingerprint {
		return errFingerprint
	}

	return nil
}

// join joins the controller with the join string of the config: it sends
// the token, who the agent is and its two public keys, signed with its key,
// to the controller that the string pins, and keeps the controller's key
// that it answers. The controller admits the machine at once.
func (a *Agent) join(
	ctx context.Context,
) error {
	j, err := parseJoin(a.appConfig.Agent.Join)
	if err != nil {
		return err
	}
	req := pki.JoinRequest{
		Token:         j.token,
		MachineID:     a.machineID,
		Hostname:      a.currentHostname(),
		PublicKey:     a.pkiManager.PublicKey(),
		EncryptionKey: a.pkiManager.EncryptionPublicKey(),
	}
	req.Sign(a.pkiManager.PrivateKey())

	c := client.New("https://"+j.host, "", client.WithHTTPTransport(pinned(j.fingerprint)))
	resp, err := c.Agent.Join(ctx, client.AgentJoinInput{
		Token:         j.token,
		MachineID:     req.MachineID,
		Hostname:      req.Hostname,
		PublicKey:     base64.StdEncoding.EncodeToString(req.PublicKey),
		EncryptionKey: req.EncryptionKey,
		Signature:     base64.StdEncoding.EncodeToString(req.Signature),
	})
	if err != nil {
		return fmt.Errorf("join %s: %w", j.host, err)
	}
	key, err := base64.StdEncoding.DecodeString(resp.Data.ControllerPublicKey)
	if err != nil {
		return fmt.Errorf("the controller's key: %w", err)
	}

	return a.keepControllerKey(key)
}
