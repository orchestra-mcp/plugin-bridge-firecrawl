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

// --- scrape_url ---

// ScrapeURLSchema returns the JSON Schema for the scrape_url tool.
func ScrapeURLSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to scrape",
			},
			"formats": map[string]any{
				"type":        "array",
				"description": "Output formats (e.g., markdown, html, rawHtml, links, screenshot). Default: [\"markdown\"]",
				"items": map[string]any{
					"type": "string",
				},
			},
			"only_main_content": map[string]any{
				"type":        "boolean",
				"description": "Only return the main content of the page, excluding headers, navs, footers. Default: true",
			},
		},
		"required": []any{"url"},
	})
	return s
}

// scrapeRequest is the JSON body for POST /v1/scrape.
type scrapeRequest struct {
	URL             string   `json:"url"`
	Formats         []string `json:"formats,omitempty"`
	OnlyMainContent *bool    `json:"onlyMainContent,omitempty"`
}

// scrapeResponse is the JSON response from POST /v1/scrape.
type scrapeResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Markdown string            `json:"markdown"`
		HTML     string            `json:"html"`
		RawHTML  string            `json:"rawHtml"`
		Links    []string          `json:"links"`
		Metadata map[string]any `json:"metadata"`
	} `json:"data"`
}

// ScrapeURL returns a tool handler that scrapes a single URL via Firecrawl.
func ScrapeURL(bridge *Bridge) plugin.ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "url"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		url := helpers.GetString(req.Arguments, "url")
		formats := helpers.GetStringSlice(req.Arguments, "formats")
		if len(formats) == 0 {
			formats = []string{"markdown"}
		}

		// Default only_main_content to true. GetBool returns false for missing
		// keys, so we check whether the field is explicitly present.
		onlyMainContent := true
		if req.Arguments != nil {
			if _, exists := req.Arguments.Fields["only_main_content"]; exists {
				onlyMainContent = helpers.GetBool(req.Arguments, "only_main_content")
			}
		}

		body := scrapeRequest{
			URL:             url,
			Formats:         formats,
			OnlyMainContent: &onlyMainContent,
		}

		var resp scrapeResponse
		if err := bridge.Client.Post(ctx, "/v1/scrape", body, &resp); err != nil {
			return helpers.ErrorResult("firecrawl_error", err.Error()), nil
		}

		return helpers.TextResult(formatScrapeResult(url, &resp)), nil
	}
}

// formatScrapeResult formats the scrape response as Markdown.
func formatScrapeResult(url string, resp *scrapeResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Scraped: %s\n\n", url)

	if resp.Data.Markdown != "" {
		b.WriteString(resp.Data.Markdown)
	} else if resp.Data.HTML != "" {
		fmt.Fprintf(&b, "```html\n%s\n```\n", resp.Data.HTML)
	}

	if len(resp.Data.Links) > 0 {
		b.WriteString("\n\n---\n## Links Found\n")
		for _, link := range resp.Data.Links {
			fmt.Fprintf(&b, "- %s\n", link)
		}
	}

	if len(resp.Data.Metadata) > 0 {
		b.WriteString("\n---\n## Metadata\n")
		for k, v := range resp.Data.Metadata {
			fmt.Fprintf(&b, "- **%s:** %v\n", k, v)
		}
	}

	return b.String()
}
