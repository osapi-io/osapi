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

// Package natsauth makes the authentication of a NATS setup for osapi: a
// credential for each osapi role, with the least permissions of the role,
// and the server's part. osapi makes each secret, and writes each one to a
// file of mode 0600. Nothing takes a secret on the command line.
package natsauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"

	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/natsjwt"
)

// OsapiConfig is the file of the osapi config that a create writes: the
// part of the config that uses the credentials, to merge into osapi.yaml.
const OsapiConfig = "osapi-nats.yaml"

// File is a file that a create wrote.
type File struct {
	Role string
	Path string
}

// Result is what a create wrote, and the server's part: the config that a
// NATS server that osapi does not run needs.
type Result struct {
	Files  []File
	Server string
}

// staticRoles are the roles of a user or a credential that osapi keeps:
// the join's bootstrap user is static only with passwords.
var staticRoles = []bus.Role{bus.RoleSetup, bus.RoleController, bus.RoleAgent}

// randomRead reads random bytes. Tests swap it.
var randomRead = rand.Read

// password is a new random password.
func password() (string, error) {
	b := make([]byte, 24)
	if _, err := randomRead(b); err != nil {
		return "", fmt.Errorf("make a password: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// write writes data to name in dir, 0600, and adds it to r.
func (r *Result) write(
	ctx context.Context,
	fs avfs.VFS,
	dir string,
	role string,
	name string,
	data []byte,
) error {
	path := filepath.Join(dir, name)
	if err := atomicfile.Write(ctx, fs, path, data, 0o600); err != nil {
		return err
	}
	r.Files = append(r.Files, File{Role: role, Path: path})

	return nil
}

// Password makes a user and a random password for each osapi role, with
// the join's bootstrap user: the users of osapi's embedded server, and the
// authorization block of a server that osapi does not run.
func Password(
	ctx context.Context,
	fs avfs.VFS,
	dir string,
	namespace string,
	cfg config.NATS,
) (Result, error) {
	roles := append(append([]bus.Role{}, staticRoles...), bus.RoleJoin)
	passwords := map[bus.Role]string{}
	for _, role := range roles {
		pw, err := password()
		if err != nil {
			return Result{}, err
		}
		passwords[role] = pw
	}

	var y strings.Builder
	y.WriteString("# The NATS users of osapi, one for each role. Merge it into osapi.yaml.\n")
	y.WriteString("nats:\n  server:\n    auth:\n      type: user_pass\n      users:\n")
	for _, role := range roles {
		fmt.Fprintf(&y, "        - username: %s\n          password: %s\n          role: %s\n",
			role, passwords[role], role)
	}
	fmt.Fprintf(
		&y,
		"controller:\n  nats:\n    auth:\n      type: user_pass\n      username: %s\n      password: %s\n",
		bus.RoleController,
		passwords[bus.RoleController],
	)
	fmt.Fprintf(
		&y,
		"agent:\n  nats:\n    auth:\n      type: user_pass\n      username: %s\n      password: %s\n",
		bus.RoleAgent,
		passwords[bus.RoleAgent],
	)

	var srv strings.Builder
	srv.WriteString(
		"# The users of osapi, for the config of a NATS server that osapi does not run.\n",
	)
	srv.WriteString("authorization {\n  users = [\n")
	for _, role := range roles {
		// Each role is one of osapi's, so it has permissions.
		p, _ := bus.RolePermissions(role, namespace, cfg)
		fmt.Fprintf(&srv, "    { user: %q, password: %q, permissions: %s }\n",
			string(role), passwords[role], serverPermissions(p))
	}
	srv.WriteString("  ]\n}\n")

	var r Result
	if err := r.write(ctx, fs, dir, "osapi", OsapiConfig, []byte(y.String())); err != nil {
		return Result{}, err
	}
	if err := r.write(
		ctx,
		fs,
		dir,
		"server",
		"nats-server.conf",
		[]byte(srv.String()),
	); err != nil {
		return Result{}, err
	}
	r.Server = srv.String()

	return r, nil
}

// serverPermissions is p in the config of a NATS server.
func serverPermissions(
	p bus.Permissions,
) string {
	list := func(subjects []string) string {
		q := make([]string, 0, len(subjects))
		for _, s := range subjects {
			q = append(q, fmt.Sprintf("%q", s))
		}

		return "[" + strings.Join(q, ", ") + "]"
	}
	out := fmt.Sprintf("{ publish: { allow: %s }, subscribe: { allow: %s }",
		list(p.Publish), list(p.Subscribe))
	if p.Responses {
		out += ", allow_responses: { max: 1 }"
	}

	return out + " }"
}

// Operator makes an operator, its system account, and an account for
// osapi with JetStream and a signing key: the resolver file of the server,
// a credentials file for each osapi role, and the signing key that the
// controller mints the users of a join with.
func Operator(
	ctx context.Context,
	fs avfs.VFS,
	dir string,
	namespace string,
	cfg config.NATS,
) (Result, error) {
	keys, err := newKeys(4)
	if err != nil {
		return Result{}, err
	}
	operator, system, account, signing := keys[0], keys[1], keys[2], keys[3]
	opPub, _ := operator.PublicKey()
	sysPub, _ := system.PublicKey()
	acctPub, _ := account.PublicKey()
	signPub, _ := signing.PublicKey()

	oc := jwt.NewOperatorClaims(opPub)
	oc.Name = "osapi"
	oc.SystemAccount = sysPub
	opJWT, _ := oc.Encode(operator)

	sc := jwt.NewAccountClaims(sysPub)
	sc.Name = "SYS"
	sysJWT, _ := sc.Encode(operator)

	ac := jwt.NewAccountClaims(acctPub)
	ac.Name = "osapi"
	ac.Limits.JetStreamLimits = jwt.JetStreamLimits{
		MemoryStorage: jwt.NoLimit, DiskStorage: jwt.NoLimit,
		Streams: jwt.NoLimit, Consumer: jwt.NoLimit,
	}
	ac.SigningKeys.Add(signPub)
	acctJWT, _ := ac.Encode(operator)

	resolver := fmt.Sprintf(
		"# The operator of osapi, for a NATS server in operator mode.\n"+
			"operator: %s\nsystem_account: %s\nresolver: MEMORY\nresolver_preload: {\n  %s: %s\n  %s: %s\n}\n",
		opJWT, sysPub, sysPub, sysJWT, acctPub, acctJWT,
	)

	var r Result
	signer := natsjwt.NewSigner(signing, acctPub)
	for _, role := range staticRoles {
		p, _ := bus.RolePermissions(role, namespace, cfg)
		creds, err := newCreds(signer, "osapi-"+string(role), p, time0)
		if err != nil {
			return Result{}, err
		}
		if err := r.write(ctx, fs, dir, string(role), string(role)+".creds", creds); err != nil {
			return Result{}, err
		}
	}
	opSeed, _ := operator.Seed()
	signSeed, _ := signing.Seed()
	for _, f := range []struct {
		role, name string
		data       []byte
	}{
		{"server", "resolver.conf", []byte(resolver)},
		{"operator", "operator.nk", opSeed},
		{"controller", "signing.nk", signSeed},
	} {
		if err := r.write(ctx, fs, dir, f.role, f.name, f.data); err != nil {
			return Result{}, err
		}
	}

	y := fmt.Sprintf(
		"# The NATS auth of osapi in operator mode. Merge it into osapi.yaml.\n"+
			"nats:\n  server:\n    auth:\n      type: operator\n      resolver_file: %s\n      creds_file: %s\n"+
			"controller:\n  nats:\n    auth:\n      type: creds\n      creds_file: %s\n"+
			"  join:\n    signing_key_file: %s\n    account: %s\n"+
			"agent:\n  nats:\n    auth:\n      type: creds\n      creds_file: %s\n",
		filepath.Join(dir, "resolver.conf"), filepath.Join(dir, "setup.creds"),
		filepath.Join(dir, "controller.creds"), filepath.Join(dir, "signing.nk"), acctPub,
		filepath.Join(dir, "agent.creds"),
	)
	if err := r.write(ctx, fs, dir, "osapi", OsapiConfig, []byte(y)); err != nil {
		return Result{}, err
	}
	sort.SliceStable(r.Files, func(i, j int) bool { return r.Files[i].Path < r.Files[j].Path })
	r.Server = resolver

	return r, nil
}

// newCreds mints a user's credentials. Tests swap it.
var newCreds = (*natsjwt.Signer).NewCreds

// time0 is no expiry: the users of a create are the deployment's own.
var time0 = timeZero()

// createOperator and createAccount make keys. Tests swap them.
var (
	createOperator = nkeys.CreateOperator
	createAccount  = nkeys.CreateAccount
)

// newKeys makes n keys: an operator key, then account keys. Tests swap it.
var newKeys = func(
	n int,
) ([]nkeys.KeyPair, error) {
	out := make([]nkeys.KeyPair, 0, n)
	op, err := createOperator()
	if err != nil {
		return nil, err
	}
	out = append(out, op)
	for len(out) < n {
		kp, err := createAccount()
		if err != nil {
			return nil, err
		}
		out = append(out, kp)
	}

	return out, nil
}

// timeZero is the zero time.
func timeZero() time.Time { return time.Time{} }

// createUser makes an nkey of a user. Tests swap it.
var createUser = nkeys.CreateUser

// NKey makes an nkey for each osapi role that keeps a credential: setup,
// controller and agent. It writes the seed of each, osapi-nats.yaml with the
// nkeys of the embedded server and their roles, and the authorization block
// of a NATS server that osapi does not run. The join's bootstrap user is not
// an nkey: a join string carries a password or a JWT.
func NKey(
	ctx context.Context,
	fs avfs.VFS,
	dir string,
	namespace string,
	cfg config.NATS,
) (Result, error) {
	var r Result
	pubs := map[bus.Role]string{}
	for _, role := range staticRoles {
		kp, err := createUser()
		if err != nil {
			return Result{}, fmt.Errorf("make the nkey of %s: %w", role, err)
		}
		seed, _ := kp.Seed()
		pubs[role], _ = kp.PublicKey()
		if err := r.write(ctx, fs, dir, string(role), string(role)+".nk", seed); err != nil {
			return Result{}, err
		}
	}

	var y strings.Builder
	y.WriteString("# The nkeys of osapi, one for each role. Merge it into osapi.yaml.\n")
	fmt.Fprintf(
		&y,
		"nats:\n  server:\n    auth:\n      type: nkey\n      nkey_file: %s\n      nkeys:\n",
		filepath.Join(dir, "setup.nk"),
	)
	for _, role := range staticRoles {
		fmt.Fprintf(&y, "        - %s\n", pubs[role])
	}
	y.WriteString("      nkey_roles:\n")
	for _, role := range staticRoles {
		fmt.Fprintf(&y, "        %s: %s\n", pubs[role], role)
	}
	fmt.Fprintf(&y, "controller:\n  nats:\n    auth:\n      type: nkey\n      nkey_file: %s\n",
		filepath.Join(dir, "controller.nk"))
	fmt.Fprintf(&y, "agent:\n  nats:\n    auth:\n      type: nkey\n      nkey_file: %s\n",
		filepath.Join(dir, "agent.nk"))

	var srv strings.Builder
	srv.WriteString(
		"# The nkeys of osapi, for the config of a NATS server that osapi does not run.\n",
	)
	srv.WriteString("authorization {\n  users = [\n")
	for _, role := range staticRoles {
		p, _ := bus.RolePermissions(role, namespace, cfg)
		fmt.Fprintf(&srv, "    { nkey: %q, permissions: %s }\n", pubs[role], serverPermissions(p))
	}
	srv.WriteString("  ]\n}\n")

	if err := r.write(ctx, fs, dir, "osapi", OsapiConfig, []byte(y.String())); err != nil {
		return Result{}, err
	}
	if err := r.write(
		ctx,
		fs,
		dir,
		"server",
		"nats-server.conf",
		[]byte(srv.String()),
	); err != nil {
		return Result{}, err
	}
	r.Server = srv.String()

	return r, nil
}
