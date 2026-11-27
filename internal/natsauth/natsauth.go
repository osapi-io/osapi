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
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	iofs "io/fs"
	"path/filepath"
	"regexp"
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
	"github.com/osapi-io/osapi/internal/natsrevoke"
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

// AccountPubFile is the file of the account's public key (A...), beside
// the signing keys, that Operator writes.
const AccountPubFile = "account.pub"

// AccountJWTFile is the file of the account's JWT, as the resolver file
// preloads it, that Operator writes: the controller adds the revocations
// of the NATS users of the machines that it removes to it.
const AccountJWTFile = "account.jwt"

// The files of the revocation of an agent's NATS user, beside the others
// of Operator: the operator signing key, which signs the account again
// with its revocations, the system account's signing key, and the system
// account user that sends the account to the server.
const (
	OperatorSigningKeyFile = "operator-signing.nk"
	SystemSigningKeyFile   = "system-signing.nk"
	SystemCredsFile        = "system.creds"
)

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
	keys, err := newKeys(6)
	if err != nil {
		return Result{}, err
	}
	opSigning, err := createOperator()
	if err != nil {
		return Result{}, err
	}
	sysPub, _ := keys[1].PublicKey()

	return writeOperator(ctx, fs, dir, namespace, cfg, operatorKeys{
		operator: keys[0], systemPub: sysPub, account: keys[2], agent: keys[3], join: keys[4],
		systemSigning: keys[5], operatorSigning: opSigning,
	})
}

// operatorKeys are the keys of an operator setup: the operator, the public
// key of its system account, the account of osapi, the two scoped signing
// keys of a join, and the signing keys of the operator and of the system
// account that the revocation of an agent's NATS user needs.
type operatorKeys struct {
	operator        nkeys.KeyPair
	systemPub       string
	account         nkeys.KeyPair
	agent           nkeys.KeyPair
	join            nkeys.KeyPair
	operatorSigning nkeys.KeyPair
	systemSigning   nkeys.KeyPair
}

// UpdateOperator signs the account of osapi and the credentials of each
// osapi role again, in dir, with the permissions that each role has now and
// with the keys that Operator made there. A version of osapi that adds a
// permission to a role needs it: the account's JWT holds the agent's
// permissions as the template of its scoped signing key, and each
// credentials file holds the permissions of its role. The keys do not
// change, so each agent keeps its user and no agent joins again. A setup
// with no operator signing key or no signing key of the system account,
// from before the revocation of an agent's NATS user, gets them. The NATS
// server reads the new resolver file at its next start.
func UpdateOperator(
	ctx context.Context,
	fs avfs.VFS,
	dir string,
	namespace string,
	cfg config.NATS,
) (Result, error) {
	var k operatorKeys
	for _, f := range []struct {
		name string
		to   *nkeys.KeyPair
	}{
		{"operator.nk", &k.operator},
		{"account.nk", &k.account},
		{"agent-signing.nk", &k.agent},
		{"join-signing.nk", &k.join},
	} {
		seed, err := fs.ReadFile(filepath.Join(dir, f.name))
		if err != nil {
			return Result{}, fmt.Errorf(
				"read %s: %w; update needs the keys that operator create made",
				f.name,
				err,
			)
		}
		kp, err := nkeys.FromSeed(bytes.TrimSpace(seed))
		if err != nil {
			return Result{}, fmt.Errorf("%s: %w", f.name, err)
		}
		*f.to = kp
	}
	for _, f := range []struct {
		name string
		to   *nkeys.KeyPair
		make func() (nkeys.KeyPair, error)
	}{
		{OperatorSigningKeyFile, &k.operatorSigning, createOperator},
		{SystemSigningKeyFile, &k.systemSigning, createAccount},
	} {
		kp, err := readOrMake(fs, filepath.Join(dir, f.name), f.make)
		if err != nil {
			return Result{}, fmt.Errorf("%s: %w", f.name, err)
		}
		*f.to = kp
	}
	resolver, err := fs.ReadFile(filepath.Join(dir, "resolver.conf"))
	if err != nil {
		return Result{}, fmt.Errorf("read resolver.conf: %w", err)
	}
	m := systemAccount.FindSubmatch(resolver)
	if m == nil {
		return Result{}, errors.New("resolver.conf names no system_account")
	}
	k.systemPub = string(m[1])

	return writeOperator(ctx, fs, dir, namespace, cfg, k)
}

// readOrMake is the key of the seed in path, or a new key from newKey when
// path does not exist.
func readOrMake(
	fs avfs.VFS,
	path string,
	newKey func() (nkeys.KeyPair, error),
) (nkeys.KeyPair, error) {
	seed, err := fs.ReadFile(path)
	if errors.Is(err, iofs.ErrNotExist) {
		return newKey()
	}
	if err != nil {
		return nil, err
	}

	return nkeys.FromSeed(bytes.TrimSpace(seed))
}

// systemAccount is the system account's line of a resolver file.
var systemAccount = regexp.MustCompile(`(?m)^system_account: (A[A-Z0-9]+)$`)

// writeOperator writes the files of an operator setup with k to dir.
func writeOperator(
	ctx context.Context,
	fs avfs.VFS,
	dir string,
	namespace string,
	cfg config.NATS,
	k operatorKeys,
) (Result, error) {
	operator, account, agentKey, joinKey := k.operator, k.account, k.agent, k.join
	opPub, _ := operator.PublicKey()
	sysPub := k.systemPub
	acctPub, _ := account.PublicKey()

	opSigningPub, _ := k.operatorSigning.PublicKey()
	sysSigningPub, _ := k.systemSigning.PublicKey()

	oc := jwt.NewOperatorClaims(opPub)
	oc.Name = "osapi"
	oc.SystemAccount = sysPub
	// The controller signs osapi's account again with this key, to add the
	// revocation of an agent's NATS user: the operator's own key stays with
	// the administrator.
	oc.SigningKeys.Add(opSigningPub)
	opJWT, _ := oc.Encode(operator)

	sc := jwt.NewAccountClaims(sysPub)
	sc.Name = "SYS"
	sc.SigningKeys.Add(sysSigningPub)
	sysJWT, _ := sc.Encode(operator)

	ac := jwt.NewAccountClaims(acctPub)
	ac.Name = "osapi"
	ac.Limits.JetStreamLimits = jwt.JetStreamLimits{
		MemoryStorage: jwt.NoLimit, DiskStorage: jwt.NoLimit,
		Streams: jwt.NoLimit, Consumer: jwt.NoLimit,
	}
	// The controller keeps only two scoped signing keys: the template of
	// each is the permissions of its role, so the controller cannot mint a
	// user with more than an agent or a join user has.
	for _, k := range []struct {
		key  nkeys.KeyPair
		role bus.Role
	}{{agentKey, bus.RoleAgent}, {joinKey, bus.RoleJoin}} {
		p, _ := bus.RolePermissions(k.role, namespace, cfg)
		scope := jwt.NewUserScope()
		scope.Key, _ = k.key.PublicKey()
		scope.Role = string(k.role)
		scope.Template.Pub.Allow = p.Publish
		scope.Template.Sub.Allow = p.Subscribe
		ac.SigningKeys.AddScopedSigner(scope)
	}
	acctJWT, _ := ac.Encode(operator)

	resolver := fmt.Sprintf(
		"# The operator of osapi, for a NATS server in operator mode.\n"+
			"operator: %s\nsystem_account: %s\nresolver: MEMORY\nresolver_preload: {\n  %s: %s\n  %s: %s\n}\n",
		opJWT, sysPub, sysPub, sysJWT, acctPub, acctJWT,
	)

	var r Result
	// The static users are signed with the account's own key, which the
	// controller does not get.
	signer := natsjwt.NewSigner(account, acctPub)
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
	// The system account's user of the controller can send only the new
	// JWT of osapi's account, and take the reply.
	sysCreds, err := newCreds(
		natsjwt.NewSigner(k.systemSigning, sysPub),
		"osapi-revocation",
		bus.Permissions{
			Publish:   []string{natsrevoke.ClaimsUpdateSubject(acctPub)},
			Subscribe: []string{"_INBOX.>"},
		},
		time0,
	)
	if err != nil {
		return Result{}, err
	}
	if err := r.write(ctx, fs, dir, "controller", SystemCredsFile, sysCreds); err != nil {
		return Result{}, err
	}
	opSeed, _ := operator.Seed()
	acctSeed, _ := account.Seed()
	agentSeed, _ := agentKey.Seed()
	joinSeed, _ := joinKey.Seed()
	opSigningSeed, _ := k.operatorSigning.Seed()
	sysSigningSeed, _ := k.systemSigning.Seed()
	for _, f := range []struct {
		role, name string
		data       []byte
	}{
		{"server", "resolver.conf", []byte(resolver)},
		{"operator", "operator.nk", opSeed},
		{"operator", "account.nk", acctSeed},
		// The account's public key, which the controller reads when
		// controller.join.account is empty: it is not secret.
		{"controller", AccountPubFile, []byte(acctPub)},
		{"controller", "agent-signing.nk", agentSeed},
		{"controller", "join-signing.nk", joinSeed},
		{"controller", AccountJWTFile, []byte(acctJWT)},
		{"controller", OperatorSigningKeyFile, opSigningSeed},
		{"operator", SystemSigningKeyFile, sysSigningSeed},
	} {
		if err := r.write(ctx, fs, dir, f.role, f.name, f.data); err != nil {
			return Result{}, err
		}
	}

	y := fmt.Sprintf(
		"# The NATS auth of osapi in operator mode. Merge it into osapi.yaml.\n"+
			"nats:\n  server:\n    auth:\n      type: operator\n      resolver_file: %s\n      creds_file: %s\n"+
			"controller:\n  nats:\n    auth:\n      type: creds\n      creds_file: %s\n"+
			"  join:\n    scoped: true\n    signing_key_file: %s\n    join_signing_key_file: %s\n    account: %s\n"+
			"    revocation:\n      account_jwt_file: %s\n      operator_signing_key_file: %s\n      system_creds_file: %s\n"+
			"agent:\n  nats:\n    auth:\n      type: creds\n      creds_file: %s\n",
		filepath.Join(dir, "resolver.conf"), filepath.Join(dir, "setup.creds"),
		filepath.Join(dir, "controller.creds"), filepath.Join(dir, "agent-signing.nk"),
		filepath.Join(dir, "join-signing.nk"), acctPub,
		filepath.Join(dir, AccountJWTFile), filepath.Join(dir, OperatorSigningKeyFile),
		filepath.Join(dir, SystemCredsFile),
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
