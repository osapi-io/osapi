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

package controllercmd

import (
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/secret/vault"
)

// Defaults are the defaults of the controller section of the config, which
// controller start owns: a key that the config file and the environment do
// not set takes its value from here. osapi config default get prints them.
var Defaults = config.Controller{
	Client: config.Client{URL: "https://localhost:8080"},
	API: config.APIServer{
		Port:       8080,
		JobTimeout: "30s",
		// TLS on, with a self-signed certificate made at the first start
		// (osapi-io/osapi#1413).
		TLS: config.ServerTLS{Enabled: true},
	},
	NATS: config.NATSConnection{
		Host:       "localhost",
		Port:       4222,
		ClientName: "osapi-controller",
		Namespace:  "osapi",
		Auth:       config.NATSAuth{Type: "none"},
		TLS:        config.ClientTLS{Enabled: true},
	},
	Metrics: config.MetricsServer{
		Enabled: true, Host: "0.0.0.0", Port: 9090, TLS: config.ServerTLS{Enabled: true},
	},
	Notifications: config.NotificationsConfig{
		Enabled:          true,
		Notifier:         "log",
		RenotifyInterval: "5m",
	},
	UI: config.UIConfig{Enabled: true},
	// The callout's own connection, which only a server with auth callout
	// uses.
	Join: config.ControllerJoin{
		Callout: config.JoinCallout{Auth: config.NATSAuth{Type: "none"}},
	},
	PKI: config.ControllerPKI{
		KeyDir:              "/etc/osapi/pki",
		RotationGracePeriod: "24h",
	},
	// The vault that secret://default/<key> names. The controller makes its
	// key at its first start.
	Secrets: config.ControllerSecrets{Vaults: []vault.Spec{{
		Name:    "default",
		Type:    "nats",
		KeyFile: config.DefaultVaultKeyFile,
	}}},
}
