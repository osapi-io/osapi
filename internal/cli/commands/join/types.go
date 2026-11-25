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

package join

import "time"

// CreateInput is the input of token create.
type CreateInput struct {
	Expires string            `json:"expires" validate:"required,go_duration" doc:"How long the token admits agents, as a Go duration such as 1h"       default:"1h"`
	Uses    int               `json:"uses"    validate:"min=1"                doc:"How many agents the token admits"                                    default:"1"`
	Labels  map[string]string `json:"labels"  validate:"omitempty"            doc:"Labels that each agent that joins with the token gets, as key=value"`
	// Transport is how the agent joins: over NATS, or over HTTPS to the API.
	Transport string `json:"transport" validate:"required,oneof=nats https" doc:"How the agent joins: nats, over the NATS server, or https, over the controller's API" default:"nats"`
}

// CreateOutput is a new token: printed one time, with its join string.
type CreateOutput struct {
	_         struct{}          `title:"Join token"`
	ID        string            `                   json:"id"`
	Token     string            `                   json:"token"`
	Join      string            `                   json:"join"             label:"Join string"`
	ExpiresAt time.Time         `                   json:"expires_at"       label:"Expires"     render:"time"`
	Uses      int               `                   json:"uses"`
	Labels    map[string]string `                   json:"labels,omitempty"`
}

// ListInput is the input of token list, which has none.
type ListInput struct{}

// Row is one token of list, with no secret.
type Row struct {
	ID        string            `json:"id"`
	CreatedAt time.Time         `json:"created_at"       label:"Created"   render:"age"`
	ExpiresAt time.Time         `json:"expires_at"       label:"Expires"   render:"time"`
	UsesLeft  int               `json:"uses_left"        label:"Uses left"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// ListOutput is each join token.
type ListOutput struct {
	_     struct{} `title:"Join tokens"`
	Items []Row    `                    json:"items" empty:"No join token admits an agent."`
}

// RevokeInput names the token to revoke.
type RevokeInput struct {
	ID string `json:"id" validate:"required" doc:"The token, as token list shows it"`
}

// RevokeOutput is the revoked token.
type RevokeOutput struct {
	_  struct{} `title:"Join token"`
	ID string   `                   json:"id"`
}
