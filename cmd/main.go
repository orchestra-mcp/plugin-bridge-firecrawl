// Command bridge-firecrawl is the entry point for the bridge.firecrawl plugin
// binary. It bridges Orchestra MCP to the Firecrawl web scraping REST API.
// This plugin does NOT require storage — all operations are stateless HTTP
// calls to the Firecrawl API.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra-mcp/plugin-bridge-firecrawl/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

func main() {
	builder := plugin.New("bridge.firecrawl").
		Version("0.1.0").
		Description("Firecrawl web scraping bridge plugin").
		Author("Orchestra").
		Binary("bridge-firecrawl")

	// Build env map from OS environment variables.
	env := map[string]string{
		"FIRECRAWL_API_KEY":  os.Getenv("FIRECRAWL_API_KEY"),
		"FIRECRAWL_BASE_URL": os.Getenv("FIRECRAWL_BASE_URL"),
	}

	bp := internal.NewBridgePlugin(env)
	bp.RegisterTools(builder)

	p := builder.BuildWithTools()
	p.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := p.Run(ctx); err != nil {
		log.Fatalf("bridge.firecrawl: %v", err)
	}
}
