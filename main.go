package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/kumahq/kuma/pkg/plugins/externalpolicy"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "plugin",
		Output:     os.Stderr,
		Level:      hclog.Trace,
		JSONFormat: true,
	})
	impl := TapPolicyPlugin{
		logger: logger,
	}

	var pluginMap = map[string]plugin.Plugin{
		"externalPolicyPlugin": &externalpolicy.ExternalPolicyGoPlugin{Impl: impl},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
