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

package client

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// UserService provides user account management operations.
type UserService struct {
	client *gen.ClientWithResponses
}

// List lists all user accounts on the target host.
func (s *UserService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[UserInfoResult]], error) {
	resp, err := s.client.GetNodeUserWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("user list: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(userInfoCollectionFromList(resp.JSON200), resp.Body), nil
}

// Get retrieves a single user account by name on the target host.
func (s *UserService) Get(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[UserInfoResult]], error) {
	resp, err := s.client.GetNodeUserByNameWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("user get: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(userInfoCollectionFromGet(resp.JSON200), resp.Body), nil
}

// Create creates a user account on the target host.
func (s *UserService) Create(
	ctx context.Context,
	hostname string,
	opts UserCreateOpts,
) (*Response[Collection[UserMutationResult]], error) {
	body := gen.UserCreateRequest{
		Name: opts.Name,
	}

	if opts.UID != 0 {
		body.Uid = &opts.UID
	}

	if opts.GID != 0 {
		body.Gid = &opts.GID
	}

	if opts.Home != "" {
		body.Home = &opts.Home
	}

	if opts.Shell != "" {
		body.Shell = &opts.Shell
	}

	if opts.Groups != nil {
		body.Groups = &opts.Groups
	}

	if opts.Password != "" {
		body.Password = &opts.Password
	}

	if opts.System {
		body.System = &opts.System
	}

	resp, err := s.client.PostNodeUserWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("user create: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(userMutationCollectionFromCreate(resp.JSON200), resp.Body), nil
}

// Update updates a user account on the target host.
func (s *UserService) Update(
	ctx context.Context,
	hostname string,
	name string,
	opts UserUpdateOpts,
) (*Response[Collection[UserMutationResult]], error) {
	body := gen.UserUpdateRequest{}

	if opts.Shell != "" {
		body.Shell = &opts.Shell
	}

	if opts.Home != "" {
		body.Home = &opts.Home
	}

	if opts.Groups != nil {
		body.Groups = &opts.Groups
	}

	if opts.Lock != nil {
		body.Lock = opts.Lock
	}

	resp, err := s.client.PutNodeUserWithResponse(ctx, hostname, name, body)
	if err != nil {
		return nil, fmt.Errorf("user update: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(userMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

// Delete removes a user account on the target host.
func (s *UserService) Delete(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[UserMutationResult]], error) {
	resp, err := s.client.DeleteNodeUserWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("user delete: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(userMutationCollectionFromDelete(resp.JSON200), resp.Body), nil
}

// ListKeys returns SSH authorized keys for a user on the target host.
func (s *UserService) ListKeys(
	ctx context.Context,
	hostname string,
	username string,
) (*Response[Collection[SSHKeyInfoResult]], error) {
	resp, err := s.client.GetNodeUserSSHKeyWithResponse(ctx, hostname, username)
	if err != nil {
		return nil, fmt.Errorf("user list keys: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(sshKeyCollectionFromGen(resp.JSON200), resp.Body), nil
}

// AddKey adds an SSH authorized key for a user on the target host.
func (s *UserService) AddKey(
	ctx context.Context,
	hostname string,
	username string,
	opts SSHKeyAddOpts,
) (*Response[Collection[SSHKeyMutationResult]], error) {
	body := gen.SSHKeyAddRequest{
		Key: opts.Key,
	}

	resp, err := s.client.PostNodeUserSSHKeyWithResponse(ctx, hostname, username, body)
	if err != nil {
		return nil, fmt.Errorf("user add key: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(sshKeyMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// RemoveKey removes an SSH authorized key by fingerprint for a user on the
// target host.
func (s *UserService) RemoveKey(
	ctx context.Context,
	hostname string,
	username string,
	fingerprint string,
) (*Response[Collection[SSHKeyMutationResult]], error) {
	resp, err := s.client.DeleteNodeUserSSHKeyWithResponse(
		ctx, hostname, username, fingerprint,
	)
	if err != nil {
		return nil, fmt.Errorf("user remove key: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(sshKeyMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// ChangePassword changes a user's password on the target host.
func (s *UserService) ChangePassword(
	ctx context.Context,
	hostname string,
	name string,
	password string,
) (*Response[Collection[UserMutationResult]], error) {
	body := gen.UserPasswordRequest{
		Password: password,
	}

	resp, err := s.client.PostNodeUserPasswordWithResponse(ctx, hostname, name, body)
	if err != nil {
		return nil, fmt.Errorf("user change password: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(userMutationCollectionFromPassword(resp.JSON200), resp.Body), nil
}
