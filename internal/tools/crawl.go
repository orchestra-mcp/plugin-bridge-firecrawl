package tools

import (
	"context"
	"fmt"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/orchestra-mcp/sdk-go/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// --- crawl_site ---

// CrawlSiteSchema returns the JSON Schema for the crawl_site tool.
func CrawlSiteSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to start crawling from",
			},
			"max_depth": map[string]any{
				"type":        "number",
				"description": "Maximum crawl depth. Default: 2",
			},
			"limit": map[string]any{
				"type":        "number",
				"description": "Maximum number of pages to crawl. Default: 10",
			},
		},
		"required": []any{"url"},
	})
	return s
}

// crawlRequest is the JSON body for POST /v1/crawl.
type crawlRequest struct {
	URL      string `json:"url"`
	MaxDepth int    `json:"maxDepth,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// crawlResponse is the JSON response from POST /v1/crawl.
type crawlResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

// CrawlSite returns a tool handler that starts crawling a website via Firecrawl.
// This is an async operation — it returns a crawl job ID that can be checked
// with get_crawl_status.
func CrawlSite(bridge *Bridge) plugin.ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "url"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		url := helpers.GetString(req.Arguments, "url")
		maxDepth := helpers.GetInt(req.Arguments, "max_depth")
		if maxDepth == 0 {
			maxDepth = 2
		}
		limit := helpers.GetInt(req.Arguments, "limit")
		if limit == 0 {
			limit = 10
		}

		body := crawlRequest{
			URL:      url,
			MaxDepth: maxDepth,
			Limit:    limit,
		}

		var resp crawlResponse
		if err := bridge.Client.Post(ctx, "/v1/crawl", body, &resp); err != nil {
			return helpers.ErrorResult("firecrawl_error", err.Error()), nil
		}

		result := fmt.Sprintf(
			"# Crawl Started\n\n"+
				"- **URL:** %s\n"+
				"- **Crawl ID:** `%s`\n"+
				"- **Max Depth:** %d\n"+
				"- **Limit:** %d pages\n\n"+
				"Use `get_crawl_status` with crawl_id `%s` to check progress and retrieve results.",
			url, resp.ID, maxDepth, limit, resp.ID,
		)

		return helpers.TextResult(result), nil
	}
}

// --- get_crawl_status ---

// GetCrawlStatusSchema returns the JSON Schema for the get_crawl_status tool.
func GetCrawlStatusSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"crawl_id": map[string]any{
				"type":        "string",
				"description": "The crawl job ID returned by crawl_site",
			},
		},
		"required": []any{"crawl_id"},
	})
	return s
}

// crawlStatusResponse is the JSON response from GET /v1/crawl/{id}.
type crawlStatusResponse struct {
	Status    string `json:"status"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
	Data      []struct {
		Markdown string            `json:"markdown"`
		Metadata map[string]string `json:"metadata"`
	} `json:"data"`
}

// GetCrawlStatus returns a tool handler that checks a crawl job's status.
func GetCrawlStatus(bridge *Bridge) plugin.ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "crawl_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		crawlID := helpers.GetString(req.Arguments, "crawl_id")

		var resp crawlStatusResponse
		if err := bridge.Client.Get(ctx, "/v1/crawl/"+crawlID, &resp); err != nil {
			return helpers.ErrorResult("firecrawl_error", err.Error()), nil
		}

		return helpers.TextResult(formatCrawlStatus(crawlID, &resp)), nil
	}
}

// formatCrawlStatus formats the crawl status response as Markdown.
func formatCrawlStatus(crawlID string, resp *crawlStatusResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Crawl Status: `%s`\n\n", crawlID)
	fmt.Fprintf(&b, "- **Status:** %s\n", resp.Status)
	fmt.Fprintf(&b, "- **Progress:** %d / %d pages\n\n", resp.Completed, resp.Total)

	if len(resp.Data) > 0 {
		b.WriteString("## Results\n\n")
		for i, page := range resp.Data {
			title := page.Metadata["title"]
			sourceURL := page.Metadata["sourceURL"]
			if title == "" {
				title = fmt.Sprintf("Page %d", i+1)
			}
			fmt.Fprintf(&b, "### %s\n", title)
			if sourceURL != "" {
				fmt.Fprintf(&b, "**URL:** %s\n\n", sourceURL)
			}
			if page.Markdown != "" {
				// Truncate long pages to keep output manageable.
				content := page.Markdown
				if len(content) > 2000 {
					content = content[:2000] + "\n\n...(truncated)"
				}
				b.WriteString(content)
			}
			b.WriteString("\n\n---\n\n")
		}
	}

	return b.String()
}
