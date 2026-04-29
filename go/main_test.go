package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-keyword-planner-mcp/go/internal/keywordplanner"
)

// TestNewServer_RegistersTools verifies that the MCP server can be created and all
// tools can be registered without panicking. This catches invalid struct tags or
// schema-generation failures at test time rather than at runtime.
func TestNewServer_RegistersTools(_ *testing.T) {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-keyword-planner-mcp",
		Version: "test",
	}, nil)

	client := keywordplanner.NewClient("token", "id", "secret", "refresh", "123", "")

	mcp.AddTool(srv,
		&mcp.Tool{Name: "generate_keyword_ideas", Description: "test"},
		func(ctx context.Context, _ *mcp.CallToolRequest, input generateKeywordIdeasInput) (*mcp.CallToolResult, any, error) {
			return generateKeywordIdeas(ctx, client, input)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{Name: "get_historical_metrics", Description: "test"},
		func(ctx context.Context, _ *mcp.CallToolRequest, input getHistoricalMetricsInput) (*mcp.CallToolResult, any, error) {
			return getHistoricalMetrics(ctx, client, input)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{Name: "get_keyword_forecast", Description: "test"},
		func(ctx context.Context, _ *mcp.CallToolRequest, input getKeywordForecastInput) (*mcp.CallToolResult, any, error) {
			return getKeywordForecast(ctx, client, input)
		},
	)
}

func TestBuildForecastOptions_RejectsBothKeywordsAndSpecs(t *testing.T) {
	t.Parallel()
	_, err := buildForecastOptions(getKeywordForecastInput{
		Keywords:     []string{"a"},
		KeywordSpecs: []forecastKeywordSpecIn{{Text: "b"}},
	})
	if err == nil || !strings.Contains(err.Error(), "not both") {
		t.Errorf("err = %v, want 'not both'", err)
	}
}

func TestBuildForecastOptions_RejectsNeitherKeywordsNorSpecs(t *testing.T) {
	t.Parallel()
	_, err := buildForecastOptions(getKeywordForecastInput{})
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Errorf("err = %v, want 'required'", err)
	}
}

func TestBuildForecastOptions_RejectsPartialDateRange(t *testing.T) {
	t.Parallel()
	_, err := buildForecastOptions(getKeywordForecastInput{
		Keywords:  []string{"a"},
		StartDate: "2026-05-01",
	})
	if err == nil || !strings.Contains(err.Error(), "both") {
		t.Errorf("err = %v, want 'both'", err)
	}
}

func TestBuildForecastOptions_RejectsBadMatchType(t *testing.T) {
	t.Parallel()
	_, err := buildForecastOptions(getKeywordForecastInput{
		Keywords:  []string{"a"},
		MatchType: "FUZZY",
	})
	if err == nil || !strings.Contains(err.Error(), "match_type") {
		t.Errorf("err = %v, want match_type complaint", err)
	}
}

func TestBuildForecastOptions_RejectsBadNetwork(t *testing.T) {
	t.Parallel()
	_, err := buildForecastOptions(getKeywordForecastInput{
		Keywords:           []string{"a"},
		KeywordPlanNetwork: "DUCKDUCKGO",
	})
	if err == nil || !strings.Contains(err.Error(), "keyword_plan_network") {
		t.Errorf("err = %v", err)
	}
}

func TestBuildHistoricalRange_RequiresBothBoundaries(t *testing.T) {
	t.Parallel()
	_, err := buildHistoricalRange(&yearMonthInput{Year: 2024, Month: "JANUARY"}, nil)
	if err == nil || !strings.Contains(err.Error(), "both") {
		t.Errorf("err = %v", err)
	}
}

func TestBuildHistoricalRange_RejectsBadMonth(t *testing.T) {
	t.Parallel()
	_, err := buildHistoricalRange(
		&yearMonthInput{Year: 2024, Month: "Janua"},
		&yearMonthInput{Year: 2024, Month: "DECEMBER"},
	)
	if err == nil || !strings.Contains(err.Error(), "month") {
		t.Errorf("err = %v", err)
	}
}
