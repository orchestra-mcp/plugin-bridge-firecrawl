// Package tools contains the tool schemas and handler functions for the
// bridge.firecrawl plugin. Each exported function pair (Schema + Handler)
// follows the same pattern used across all Orchestra plugins.
//
// Unlike the bridge.ollama plugin, this plugin is stateless — it does not
// manage sessions. Each tool call is an independent HTTP request to the
// Firecrawl REST API.
package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-bridge-firecrawl/internal/client"
)

// ToolHandler is an alias for readability.
type ToolHandler = func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error)

// Bridge holds the injected dependencies that tool handlers need.
type Bridge struct {
	Client *client.FirecrawlClient
}
