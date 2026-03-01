package bridgefirecrawl

import (
	"github.com/orchestra-mcp/plugin-bridge-firecrawl/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all Firecrawl bridge tools to the builder.
func Register(builder *plugin.PluginBuilder, env map[string]string) {
	bp := internal.NewBridgePlugin(env)
	bp.RegisterTools(builder)
}
