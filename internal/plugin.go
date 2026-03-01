package internal

import (
	"github.com/orchestra-mcp/plugin-bridge-firecrawl/internal/client"
	"github.com/orchestra-mcp/plugin-bridge-firecrawl/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// BridgePlugin manages the Firecrawl client and registers all bridge tools.
type BridgePlugin struct {
	client *client.FirecrawlClient
}

// NewBridgePlugin creates a new BridgePlugin with a Firecrawl client
// configured from the given environment variables.
func NewBridgePlugin(env map[string]string) *BridgePlugin {
	return &BridgePlugin{
		client: client.NewFirecrawlClient(env),
	}
}

// RegisterTools registers all 5 Firecrawl tools with the plugin builder.
func (bp *BridgePlugin) RegisterTools(builder *plugin.PluginBuilder) {
	bridge := &tools.Bridge{
		Client: bp.client,
	}

	// --- Scrape tool (1) ---
	builder.RegisterTool("scrape_url",
		"Scrape a single URL and return its content as Markdown",
		tools.ScrapeURLSchema(), tools.ScrapeURL(bridge))

	// --- Crawl tools (2) ---
	builder.RegisterTool("crawl_site",
		"Start crawling a website (async). Returns a crawl ID to check status",
		tools.CrawlSiteSchema(), tools.CrawlSite(bridge))

	builder.RegisterTool("get_crawl_status",
		"Check the status of a crawl job and retrieve results when complete",
		tools.GetCrawlStatusSchema(), tools.GetCrawlStatus(bridge))

	// --- Search & map tools (2) ---
	builder.RegisterTool("search_web",
		"Search the web and return results with content via Firecrawl",
		tools.SearchWebSchema(), tools.SearchWeb(bridge))

	builder.RegisterTool("map_site",
		"Map all URLs on a website to discover its structure",
		tools.MapSiteSchema(), tools.MapSite(bridge))
}
