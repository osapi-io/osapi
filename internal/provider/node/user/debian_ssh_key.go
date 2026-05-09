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

package user

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	authorizedKeysFile = "authorized_keys"
	sshDirName         = ".ssh"
	sshDirMode         = 0o700
	authorizedKeysMode = 0o600
	minKeyFields       = 2
)

// ListKeys returns the SSH authorized keys for the given user.
func (d *Debian) ListKeys(
	ctx context.Context,
	username string,
) ([]SSHKey, error) {
	_ = ctx

	d.logger.Debug(
		"executing user.ListKeys",
		slog.String("username", username),
	)

	homeDir, err := d.userHomeDir(username)
	if err != nil {
		return nil, fmt.Errorf("ssh key: list: %w", err)
	}

	authKeysPath := filepath.Join(homeDir, sshDirName, authorizedKeysFile)

	content, err := d.fs.ReadFile(authKeysPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []SSHKey{}, nil
		}

		return nil, fmt.Errorf("ssh key: list: read %s: %w", authKeysPath, err)
	}

	keys := d.parseAuthorizedKeys(string(content))

	return keys, nil
}

// AddKey adds an SSH public key to the user's authorized_keys file.
func (d *Debian) AddKey(
	ctx context.Context,
	username string,
	key SSHKey,
) (*SSHKeyResult, error) {
	_ = ctx

	d.logger.Debug(
		"executing user.AddKey",
		slog.String("username", username),
	)

	homeDir, err := d.userHomeDir(username)
	if err != nil {
		return nil, fmt.Errorf("ssh key: add: %w", err)
	}

	sshDir := filepath.Join(homeDir, sshDirName)
	authKeysPath := filepath.Join(sshDir, authorizedKeysFile)

	// Check if key already exists by fingerprint.
	content, err := d.fs.ReadFile(authKeysPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("ssh key: add: read %s: %w", authKeysPath, err)
	}

	if err == nil {
		newFingerprint := fingerprintFromLine(key.RawLine)
		if newFingerprint != "" {
			existing := d.parseAuthorizedKeys(string(content))
			for _, k := range existing {
				if k.Fingerprint == newFingerprint {
					return &SSHKeyResult{Changed: false}, nil
				}
			}
		}
	}

	// Create .ssh directory if missing.
	if err := d.fs.MkdirAll(sshDir, sshDirMode); err != nil {
		return nil, fmt.Errorf("ssh key: add: mkdir %s: %w", sshDir, err)
	}

	// Append key to authorized_keys.
	f, err := d.fs.OpenFile(
		authKeysPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		authorizedKeysMode,
	)
	if err != nil {
		return nil, fmt.Errorf("ssh key: add: open %s: %w", authKeysPath, err)
	}

	_, writeErr := f.Write([]byte(key.RawLine + "\n"))

	if closeErr := f.Close(); closeErr != nil && writeErr == nil {
		writeErr = closeErr
	}

	if writeErr != nil {
		return nil, fmt.Errorf("ssh key: add: write %s: %w", authKeysPath, writeErr)
	}

	// Best-effort chown.
	_, chownErr := d.execManager.RunPrivilegedCmd("chown", []string{
		"-R",
		username + ":" + username,
		sshDir,
	})
	if chownErr != nil {
		d.logger.Warn(
			"chown failed for ssh directory",
			slog.String("username", username),
			slog.String("path", sshDir),
			slog.String("error", chownErr.Error()),
		)
	}

	d.logger.Info(
		"ssh key added",
		slog.String("username", username),
	)

	return &SSHKeyResult{Changed: true}, nil
}

// RemoveKey removes an SSH public key by fingerprint from the user's
// authorized_keys file.
func (d *Debian) RemoveKey(
	ctx context.Context,
	username string,
	fingerprint string,
) (*SSHKeyResult, error) {
	_ = ctx

	d.logger.Debug(
		"executing user.RemoveKey",
		slog.String("username", username),
		slog.String("fingerprint", fingerprint),
	)

	homeDir, err := d.userHomeDir(username)
	if err != nil {
		return nil, fmt.Errorf("ssh key: remove: %w", err)
	}

	authKeysPath := filepath.Join(homeDir, sshDirName, authorizedKeysFile)

	content, err := d.fs.ReadFile(authKeysPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &SSHKeyResult{Changed: false}, nil
		}

		return nil, fmt.Errorf("ssh key: remove: read %s: %w", authKeysPath, err)
	}

	lines := strings.Split(string(content), "\n")
	var remaining []string

	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			remaining = append(remaining, line)

			continue
		}

		fp := fingerprintFromLine(trimmed)
		if fp == fingerprint {
			found = true

			continue
		}

		remaining = append(remaining, line)
	}

	if !found {
		return &SSHKeyResult{Changed: false}, nil
	}

	output := strings.Join(remaining, "\n")

	if err := d.fs.WriteFile(authKeysPath, []byte(output), authorizedKeysMode); err != nil {
		return nil, fmt.Errorf("ssh key: remove: write %s: %w", authKeysPath, err)
	}

	d.logger.Info(
		"ssh key removed",
		slog.String("username", username),
		slog.String("fingerprint", fingerprint),
	)

	return &SSHKeyResult{Changed: true}, nil
}

// userHomeDir resolves a user's home directory from /etc/passwd.
func (d *Debian) userHomeDir(
	username string,
) (string, error) {
	f, err := d.fs.Open(passwdFile)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", passwdFile, err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < passwdFields {
			continue
		}

		if fields[0] == username {
			return fields[5], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read %s: %w", passwdFile, err)
	}

	return "", fmt.Errorf("user %q not found", username)
}

// parseAuthorizedKeys parses the content of an authorized_keys file into
// SSHKey entries. Lines that are empty, comments, or malformed are skipped.
func (d *Debian) parseAuthorizedKeys(
	content string,
) []SSHKey {
	var keys []SSHKey

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < minKeyFields {
			d.logger.Debug(
				"skipping malformed authorized_keys line",
				slog.String("line", trimmed),
			)

			continue
		}

		fp := computeFingerprint(fields[1])
		if fp == "" {
			d.logger.Debug(
				"skipping line with invalid base64 key data",
				slog.String("line", trimmed),
			)

			continue
		}

		key := SSHKey{
			Type:        fields[0],
			Fingerprint: fp,
		}

		if len(fields) > minKeyFields {
			key.Comment = strings.Join(fields[minKeyFields:], " ")
		}

		keys = append(keys, key)
	}

	return keys
}

// computeFingerprint computes the SHA256 fingerprint of base64-encoded key
// data, returning the OpenSSH format "SHA256:<base64-raw>".
// Returns empty string if the key data is not valid base64.
func computeFingerprint(
	keyData string,
) string {
	decoded, err := base64.StdEncoding.DecodeString(keyData)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(decoded)

	return "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])
}

// fingerprintFromLine extracts and computes the fingerprint from a full
// authorized_keys line. Returns empty string if the line is malformed or
// contains invalid key data.
func fingerprintFromLine(
	line string,
) string {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) < minKeyFields {
		return ""
	}

	return computeFingerprint(fields[1])
}
