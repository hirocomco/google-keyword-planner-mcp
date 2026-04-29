// Command google-keyword-planner-mcp is an MCP server that exposes Google Ads Keyword Planner
// as tools for AI assistants. It communicates via STDIO using the MCP protocol.
//
// Usage:
//
//	google-keyword-planner-mcp [--developer-token <token>] [--client-id <id>]
//	    [--client-secret <secret>] [--refresh-token <token>] [--customer-id <id>]
//
// Credential resolution order: CLI flags > environment variables > .env file.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-keyword-planner-mcp/go/internal/config"
	"github.com/ncosentino/google-keyword-planner-mcp/go/internal/keywordplanner"
)

var version = "dev"

var validNetworks = map[string]bool{
	"GOOGLE_SEARCH":              true,
	"GOOGLE_SEARCH_AND_PARTNERS": true,
}

var validMatchTypes = map[string]bool{
	"EXACT": true, "PHRASE": true, "BROAD": true,
}

var validMonths = map[string]bool{
	"JANUARY": true, "FEBRUARY": true, "MARCH": true, "APRIL": true,
	"MAY": true, "JUNE": true, "JULY": true, "AUGUST": true,
	"SEPTEMBER": true, "OCTOBER": true, "NOVEMBER": true, "DECEMBER": true,
}

func validateNetwork(n string) error {
	if n == "" || validNetworks[n] {
		return nil
	}
	return fmt.Errorf("keyword_plan_network must be GOOGLE_SEARCH or GOOGLE_SEARCH_AND_PARTNERS, got %q", n)
}

func validateMatchType(field, mt string) error {
	if mt == "" || validMatchTypes[mt] {
		return nil
	}
	return fmt.Errorf("%s must be EXACT, PHRASE, or BROAD, got %q", field, mt)
}

func validateYearMonth(field string, ym *yearMonthInput) error {
	if ym == nil {
		return nil
	}
	if ym.Year < 1900 || ym.Year > 9999 {
		return fmt.Errorf("%s.year out of range: %d", field, ym.Year)
	}
	if !validMonths[ym.Month] {
		return fmt.Errorf("%s.month must be one of JANUARY..DECEMBER, got %q", field, ym.Month)
	}
	return nil
}

func buildHistoricalRange(startIn, endIn *yearMonthInput) (*keywordplanner.YearMonthRange, error) {
	if startIn == nil && endIn == nil {
		return nil, nil
	}
	if (startIn == nil) != (endIn == nil) {
		return nil, fmt.Errorf("historical_metrics_start and historical_metrics_end must both be set or both omitted")
	}
	if err := validateYearMonth("historical_metrics_start", startIn); err != nil {
		return nil, err
	}
	if err := validateYearMonth("historical_metrics_end", endIn); err != nil {
		return nil, err
	}
	return &keywordplanner.YearMonthRange{
		Start: keywordplanner.YearMonth{Year: startIn.Year, Month: startIn.Month},
		End:   keywordplanner.YearMonth{Year: endIn.Year, Month: endIn.Month},
	}, nil
}

func buildKeywordIdeasOptions(input generateKeywordIdeasInput) (keywordplanner.KeywordIdeasOptions, error) {
	if err := validateNetwork(input.KeywordPlanNetwork); err != nil {
		return keywordplanner.KeywordIdeasOptions{}, err
	}
	hr, err := buildHistoricalRange(input.HistoricalDateStart, input.HistoricalDateEnd)
	if err != nil {
		return keywordplanner.KeywordIdeasOptions{}, err
	}
	return keywordplanner.KeywordIdeasOptions{
		Language:                    input.Language,
		GeoTargetConstants:          input.GeoTargetConstants,
		KeywordPlanNetwork:          input.KeywordPlanNetwork,
		IncludeAdultKeywords:        input.IncludeAdultKeywords,
		KeywordAnnotation:           input.KeywordAnnotation,
		AggregateMetrics:            input.AggregateMetrics,
		HistoricalDateRange:         hr,
		HistoricalIncludeAverageCpc: input.IncludeAverageCpc,
		CurrencyCode:                input.CurrencyCode,
		ToplevelDomain:              input.ToplevelDomain,
	}, nil
}

func buildHistoricalMetricsOptions(input getHistoricalMetricsInput) (keywordplanner.HistoricalMetricsOptions, error) {
	if err := validateNetwork(input.KeywordPlanNetwork); err != nil {
		return keywordplanner.HistoricalMetricsOptions{}, err
	}
	hr, err := buildHistoricalRange(input.HistoricalDateStart, input.HistoricalDateEnd)
	if err != nil {
		return keywordplanner.HistoricalMetricsOptions{}, err
	}
	return keywordplanner.HistoricalMetricsOptions{
		Language:                    input.Language,
		GeoTargetConstants:          input.GeoTargetConstants,
		KeywordPlanNetwork:          input.KeywordPlanNetwork,
		IncludeAdultKeywords:        input.IncludeAdultKeywords,
		AggregateMetrics:            input.AggregateMetrics,
		HistoricalDateRange:         hr,
		HistoricalIncludeAverageCpc: input.IncludeAverageCpc,
	}, nil
}

func buildForecastOptions(input getKeywordForecastInput) (keywordplanner.ForecastOptions, error) {
	if err := validateNetwork(input.KeywordPlanNetwork); err != nil {
		return keywordplanner.ForecastOptions{}, err
	}
	if err := validateMatchType("match_type", input.MatchType); err != nil {
		return keywordplanner.ForecastOptions{}, err
	}
	if (input.StartDate == "") != (input.EndDate == "") {
		return keywordplanner.ForecastOptions{}, fmt.Errorf("start_date and end_date must both be set or both omitted")
	}
	if input.StartDate != "" && input.StartDate > input.EndDate {
		return keywordplanner.ForecastOptions{}, fmt.Errorf("start_date %q is after end_date %q", input.StartDate, input.EndDate)
	}
	if len(input.Keywords) > 0 && len(input.KeywordSpecs) > 0 {
		return keywordplanner.ForecastOptions{}, fmt.Errorf("provide either keywords or keyword_specs, not both")
	}
	if len(input.Keywords) == 0 && len(input.KeywordSpecs) == 0 {
		return keywordplanner.ForecastOptions{}, fmt.Errorf("either keywords or keyword_specs is required")
	}
	specs := make([]keywordplanner.ForecastKeywordSpec, 0, len(input.KeywordSpecs))
	for i, s := range input.KeywordSpecs {
		if s.Text == "" {
			return keywordplanner.ForecastOptions{}, fmt.Errorf("keyword_specs[%d].text is required", i)
		}
		if err := validateMatchType(fmt.Sprintf("keyword_specs[%d].match_type", i), s.MatchType); err != nil {
			return keywordplanner.ForecastOptions{}, err
		}
		specs = append(specs, keywordplanner.ForecastKeywordSpec{
			Text: s.Text, MatchType: s.MatchType, MaxCpcBidMicros: s.MaxCpcBidMicros,
		})
	}
	return keywordplanner.ForecastOptions{
		GeoTargetConstants: input.GeoTargetConstants,
		LanguageConstants:  input.LanguageConstants,
		KeywordPlanNetwork: input.KeywordPlanNetwork,
		MatchType:          input.MatchType,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		KeywordSpecs:       specs,
	}, nil
}

func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling result: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func errorResult(err error) *mcp.CallToolResult {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}
}

func main() {
	developerToken := flag.String("developer-token", "", "Google Ads developer token")
	clientID := flag.String("client-id", "", "OAuth2 client ID")
	clientSecret := flag.String("client-secret", "", "OAuth2 client secret")
	refreshToken := flag.String("refresh-token", "", "OAuth2 refresh token")
	customerID := flag.String("customer-id", "", "Google Ads customer ID")
	loginCustomerID := flag.String("login-customer-id", "", "Google Ads manager/MCC account ID (required when customer-id is a sub-account)")
	flag.Parse()

	// All diagnostic output must go to stderr to avoid corrupting the MCP STDIO stream.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Resolve(config.Flags{
		DeveloperToken:  *developerToken,
		ClientID:        *clientID,
		ClientSecret:    *clientSecret,
		RefreshToken:    *refreshToken,
		CustomerID:      *customerID,
		LoginCustomerID: *loginCustomerID,
	})

	if !cfg.IsComplete() {
		slog.Error("incomplete Google Ads credentials",
			"hint", "set GOOGLE_ADS_DEVELOPER_TOKEN, GOOGLE_ADS_CLIENT_ID, "+
				"GOOGLE_ADS_CLIENT_SECRET, GOOGLE_ADS_REFRESH_TOKEN, GOOGLE_ADS_CUSTOMER_ID")
		os.Exit(1)
	}

	client := keywordplanner.NewClient(
		cfg.DeveloperToken, cfg.ClientID, cfg.ClientSecret, cfg.RefreshToken, cfg.CustomerID, cfg.LoginCustomerID,
	)

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-keyword-planner-mcp",
		Version: version,
	}, nil)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "generate_keyword_ideas",
			Description: "Generate keyword ideas from seed keywords and/or a URL using Google Ads Keyword Planner. Returns related keywords with average monthly search volume, competition level, and CPC estimates. Optional geo_target_constants narrows by location (e.g. ['geoTargetConstants/2840'] for US, ['geoTargetConstants/21137'] for California). Optional language uses 'languageConstants/1000' for English. Set historical_metrics_start/end to scope volume history.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input generateKeywordIdeasInput) (*mcp.CallToolResult, any, error) {
			return generateKeywordIdeas(ctx, client, input)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "get_historical_metrics",
			Description: "Get historical search volume and competition metrics for a list of specific keywords. Optional geo_target_constants narrows by location. Set historical_metrics_start/end (e.g. {year:2024,month:'JANUARY'}) to scope the date range; defaults to Google's most recent ~12 months.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input getHistoricalMetricsInput) (*mcp.CallToolResult, any, error) {
			return getHistoricalMetrics(ctx, client, input)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "get_keyword_forecast",
			Description: "Get projected impressions, clicks, and cost for a set of keywords at a given max CPC bid. Uses manual CPC bidding only. Pass keywords (string list, all share match_type) OR keyword_specs (per-keyword text/match_type/max_cpc_micros) — not both. Optional geo_target_constants and language_constants narrow targeting. Default match_type is BROAD; set EXACT for precise volume estimates.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input getKeywordForecastInput) (*mcp.CallToolResult, any, error) {
			return getKeywordForecast(ctx, client, input)
		},
	)

	slog.Info("google-keyword-planner-mcp starting", "version", version, "transport", "stdio")
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

// yearMonthInput is the JSON-input shape for a year/month boundary.
type yearMonthInput struct {
	Year  int32  `json:"year,omitempty"  jsonschema:"Four-digit year (e.g. 2024)."`
	Month string `json:"month,omitempty" jsonschema:"Month enum: JANUARY, FEBRUARY, MARCH, APRIL, MAY, JUNE, JULY, AUGUST, SEPTEMBER, OCTOBER, NOVEMBER, DECEMBER."`
}

// generateKeywordIdeasInput is the input schema for the generate_keyword_ideas tool.
type generateKeywordIdeasInput struct {
	SeedKeywords        []string        `json:"seed_keywords,omitempty" jsonschema:"Seed keywords to generate ideas from (e.g. ['lawn care','fertilizer']). At least one of seed_keywords or url must be provided."`
	URL                 string          `json:"url,omitempty"           jsonschema:"A URL to generate ideas from (e.g. 'https://example.com'). At least one of seed_keywords or url must be provided."`
	Language            string          `json:"language,omitempty"      jsonschema:"Language resource name (e.g. 'languageConstants/1000' for English). Omit to use all languages."`
	GeoTargetConstants  []string        `json:"geo_target_constants,omitempty" jsonschema:"Location resource names (e.g. ['geoTargetConstants/2840'] = US, ['geoTargetConstants/21137'] = California). Omit for Google's default."`
	KeywordPlanNetwork  string          `json:"keyword_plan_network,omitempty" jsonschema:"GOOGLE_SEARCH or GOOGLE_SEARCH_AND_PARTNERS. Omit for Google's default."`
	IncludeAdultKeywords bool           `json:"include_adult_keywords,omitempty" jsonschema:"Include adult-content keywords. Default false."`
	KeywordAnnotation   []string        `json:"keyword_annotation,omitempty" jsonschema:"Annotation types (e.g. ['KEYWORD_CONCEPT'])."`
	AggregateMetrics    []string        `json:"aggregate_metrics,omitempty" jsonschema:"Aggregate metric types. Currently only 'DEVICE' is supported by Google."`
	HistoricalDateStart *yearMonthInput `json:"historical_metrics_start,omitempty" jsonschema:"Start of historical metrics window. If set, historical_metrics_end must also be set."`
	HistoricalDateEnd   *yearMonthInput `json:"historical_metrics_end,omitempty" jsonschema:"End of historical metrics window. If set, historical_metrics_start must also be set."`
	IncludeAverageCpc   bool            `json:"include_average_cpc,omitempty" jsonschema:"Include average CPC in historical metrics output. Default false."`
	CurrencyCode        string          `json:"currency_code,omitempty" jsonschema:"ISO 4217 currency code (e.g. 'USD')."`
	ToplevelDomain      string          `json:"toplevel_domain,omitempty" jsonschema:"Country TLD (e.g. 'com', 'co.uk')."`
}

// getHistoricalMetricsInput is the input schema for the get_historical_metrics tool.
type getHistoricalMetricsInput struct {
	Keywords            []string        `json:"keywords,omitempty" jsonschema:"List of keywords to get historical search metrics for."`
	Language            string          `json:"language,omitempty" jsonschema:"Language resource name (e.g. 'languageConstants/1000')."`
	GeoTargetConstants  []string        `json:"geo_target_constants,omitempty" jsonschema:"Location resource names (e.g. ['geoTargetConstants/21137'] for California)."`
	KeywordPlanNetwork  string          `json:"keyword_plan_network,omitempty" jsonschema:"GOOGLE_SEARCH or GOOGLE_SEARCH_AND_PARTNERS."`
	IncludeAdultKeywords bool           `json:"include_adult_keywords,omitempty"`
	AggregateMetrics    []string        `json:"aggregate_metrics,omitempty" jsonschema:"Currently only 'DEVICE' is supported."`
	HistoricalDateStart *yearMonthInput `json:"historical_metrics_start,omitempty"`
	HistoricalDateEnd   *yearMonthInput `json:"historical_metrics_end,omitempty"`
	IncludeAverageCpc   bool            `json:"include_average_cpc,omitempty"`
}

// forecastKeywordSpecIn is the JSON-input shape for a per-keyword forecast entry.
type forecastKeywordSpecIn struct {
	Text            string `json:"text,omitempty"`
	MatchType       string `json:"match_type,omitempty"`
	MaxCpcBidMicros int64  `json:"max_cpc_micros,omitempty"`
}

// getKeywordForecastInput is the input schema for the get_keyword_forecast tool.
type getKeywordForecastInput struct {
	Keywords           []string                `json:"keywords,omitempty" jsonschema:"List of keywords to forecast. Mutually exclusive with keyword_specs."`
	KeywordSpecs       []forecastKeywordSpecIn `json:"keyword_specs,omitempty" jsonschema:"Per-keyword spec list. Mutually exclusive with keywords. Use when you need per-keyword match types or bid overrides."`
	MaxCPCMicros       int64                   `json:"max_cpc_micros,omitempty" jsonschema:"Maximum CPC bid in micros (1,000,000 = $1.00)."`
	ForecastDays       int                     `json:"forecast_days,omitempty" jsonschema:"Days to forecast. Default 30. Ignored when start_date/end_date are set."`
	GeoTargetConstants []string                `json:"geo_target_constants,omitempty" jsonschema:"Location resource names (e.g. ['geoTargetConstants/21137'] for California)."`
	LanguageConstants  []string                `json:"language_constants,omitempty" jsonschema:"Language resource names (e.g. ['languageConstants/1000'])."`
	KeywordPlanNetwork string                  `json:"keyword_plan_network,omitempty" jsonschema:"GOOGLE_SEARCH or GOOGLE_SEARCH_AND_PARTNERS."`
	MatchType          string                  `json:"match_type,omitempty" jsonschema:"Default match type (EXACT, PHRASE, BROAD). Applied to all keywords or as fallback for keyword_specs entries that omit match_type. Default BROAD."`
	StartDate          string                  `json:"start_date,omitempty" jsonschema:"YYYY-MM-DD start of forecast window. Both start_date and end_date must be set together."`
	EndDate            string                  `json:"end_date,omitempty" jsonschema:"YYYY-MM-DD end of forecast window."`
}

func generateKeywordIdeas(ctx context.Context, client *keywordplanner.Client, input generateKeywordIdeasInput) (*mcp.CallToolResult, any, error) {
	opts, err := buildKeywordIdeasOptions(input)
	if err != nil {
		return errorResult(err), nil, nil
	}
	result, err := client.GenerateKeywordIdeas(ctx, input.SeedKeywords, input.URL, opts)
	if err != nil {
		return errorResult(fmt.Errorf("generating keyword ideas: %w", err)), nil, nil
	}
	return jsonResult(result)
}

func getHistoricalMetrics(ctx context.Context, client *keywordplanner.Client, input getHistoricalMetricsInput) (*mcp.CallToolResult, any, error) {
	opts, err := buildHistoricalMetricsOptions(input)
	if err != nil {
		return errorResult(err), nil, nil
	}
	result, err := client.GetHistoricalMetrics(ctx, input.Keywords, opts)
	if err != nil {
		return errorResult(fmt.Errorf("getting historical metrics: %w", err)), nil, nil
	}
	return jsonResult(result)
}

func getKeywordForecast(ctx context.Context, client *keywordplanner.Client, input getKeywordForecastInput) (*mcp.CallToolResult, any, error) {
	opts, err := buildForecastOptions(input)
	if err != nil {
		return errorResult(err), nil, nil
	}
	result, err := client.GetKeywordForecast(ctx, input.Keywords, input.MaxCPCMicros, input.ForecastDays, opts)
	if err != nil {
		return errorResult(fmt.Errorf("getting keyword forecast: %w", err)), nil, nil
	}
	return jsonResult(result)
}
