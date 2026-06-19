package sdk

import (
	orbisplugin "example.com/mud/plugin"
	"github.com/hashicorp/go-plugin"
)

// Serve starts the game binary's plugin server. Call this from main().
func Serve(game Game) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: orbisplugin.Handshake,
		Plugins: plugin.PluginSet{
			"game": &orbisplugin.GameGRPCPlugin{
				Impl: NewAdapter(game),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
