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

// --- search_web ---

// SearchWebSchema returns the JSON Schema for the search_web tool.
func SearchWebSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query",
			},
			"limit": map[string]any{
				"type":        "number",
				"description": "Maximum number of results to return. Default: 5",
			},
		},
		"required": []any{"query"},
	})
	return s
}

// searchRequest is the JSON body for POST /v1/search.
type searchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// searchResponse is the JSON response from POST /v1/search.
type searchResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		Markdown    string `json:"markdown"`
		Description string `json:"description"`
	} `json:"data"`
}

// SearchWeb returns a tool handler that searches the web via Firecrawl.
func SearchWeb(bridge *Bridge) plugin.ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "query"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		query := helpers.GetString(req.Arguments, "query")
		limit := helpers.GetInt(req.Arguments, "limit")
		if limit == 0 {
			limit = 5
		}

		body := searchRequest{
			Query: query,
			Limit: limit,
		}

		var resp searchResponse
		if err := bridge.Client.Post(ctx, "/v1/search", body, &resp); err != nil {
			return helpers.ErrorResult("firecrawl_error", err.Error()), nil
		}

		return helpers.TextResult(formatSearchResults(query, &resp)), nil
	}
}

// formatSearchResults formats the search response as Markdown.
func formatSearchResults(query string, resp *searchResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Search Results: %s\n\n", query)

	if len(resp.Data) == 0 {
		b.WriteString("No results found.\n")
		return b.String()
	}

	for i, result := range resp.Data {
		fmt.Fprintf(&b, "## %d. %s\n", i+1, result.Title)
		fmt.Fprintf(&b, "**URL:** %s\n\n", result.URL)

		if result.Description != "" {
			fmt.Fprintf(&b, "%s\n\n", result.Description)
		}

		if result.Markdown != "" {
			content := result.Markdown
			if len(content) > 1500 {
				content = content[:1500] + "\n\n...(truncated)"
			}
			b.WriteString(content)
			b.WriteString("\n")
		}

		b.WriteString("\n---\n\n")
	}

	return b.String()
}

// --- map_site ---

// MapSiteSchema returns the JSON Schema for the map_site tool.
func MapSiteSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The website URL to map",
			},
		},
		"required": []any{"url"},
	})
	return s
}

// mapRequest is the JSON body for POST /v1/map.
type mapRequest struct {
	URL string `json:"url"`
}

// mapResponse is the JSON response from POST /v1/map.
type mapResponse struct {
	Success bool     `json:"success"`
	Links   []string `json:"links"`
}

// MapSite returns a tool handler that maps all URLs on a website via Firecrawl.
func MapSite(bridge *Bridge) plugin.ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "url"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		url := helpers.GetString(req.Arguments, "url")

		body := mapRequest{
			URL: url,
		}

		var resp mapResponse
		if err := bridge.Client.Post(ctx, "/v1/map", body, &resp); err != nil {
			return helpers.ErrorResult("firecrawl_error", err.Error()), nil
		}

		return helpers.TextResult(formatMapResults(url, &resp)), nil
	}
}

// formatMapResults formats the map response as Markdown.
func formatMapResults(url string, resp *mapResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Site Map: %s\n\n", url)
	fmt.Fprintf(&b, "**Total URLs found:** %d\n\n", len(resp.Links))

	if len(resp.Links) == 0 {
		b.WriteString("No URLs discovered.\n")
		return b.String()
	}

	for _, link := range resp.Links {
		fmt.Fprintf(&b, "- %s\n", link)
	}

	return b.String()
}
