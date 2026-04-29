// Package keywordplanner provides a client for the Google Ads Keyword Planner API.
package keywordplanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	tokenURL    = "https://oauth2.googleapis.com/token"
	adsAPIBase  = "https://googleads.googleapis.com/v23"
	adsAPIVersion = "v23"
	httpTimeout = 30 * time.Second
)

// Client calls the Google Ads Keyword Planner API.
type Client struct {
	httpClient      *http.Client
	developerToken  string
	customerID      string
	loginCustomerID string
	baseURL         string
	tokenSource     oauth2.TokenSource
}

// NewClient creates a Client with the provided OAuth2 credentials.
// loginCustomerID is the manager/MCC account ID; set it when customerID is a sub-account.
func NewClient(developerToken, clientID, clientSecret, refreshToken, customerID, loginCustomerID string) *Client {
	return NewClientWithBaseURL(developerToken, clientID, clientSecret, refreshToken, customerID, loginCustomerID, adsAPIBase)
}

// NewClientWithBaseURL creates a Client with a custom API base URL. Intended for testing.
func NewClientWithBaseURL(developerToken, clientID, clientSecret, refreshToken, customerID, loginCustomerID, baseURL string) *Client {
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: tokenURL},
	}
	token := &oauth2.Token{RefreshToken: refreshToken}
	ts := conf.TokenSource(context.Background(), token)

	base := oauth2.NewClient(context.Background(), ts)
	base.Timeout = httpTimeout

	return &Client{
		httpClient:      base,
		developerToken:  developerToken,
		customerID:      customerID,
		loginCustomerID: loginCustomerID,
		baseURL:         baseURL,
		tokenSource:     ts,
	}
}

// newTestClient creates a Client that uses a plain http.Client (no OAuth2) for unit tests.
func newTestClient(developerToken, customerID, loginCustomerID, baseURL string, httpClient *http.Client) *Client {
	return &Client{
		httpClient:      httpClient,
		developerToken:  developerToken,
		customerID:      customerID,
		loginCustomerID: loginCustomerID,
		baseURL:         baseURL,
	}
}

// NewTestClient is exported solely for use in package-level tests.
// Do not use in production code.
func NewTestClient(developerToken, customerID, loginCustomerID, baseURL string, httpClient *http.Client) *Client {
	return newTestClient(developerToken, customerID, loginCustomerID, baseURL, httpClient)
}

// GenerateKeywordIdeas returns keyword ideas for the given seed keywords and/or URL.
// opts carries optional fields (geo, language, network, etc.). Pass a zero-value
// KeywordIdeasOptions{} to keep current behavior.
func (c *Client) GenerateKeywordIdeas(
	ctx context.Context,
	seedKeywords []string,
	seedURL string,
	opts KeywordIdeasOptions,
) (*KeywordIdeasResponse, error) {
	reqBody := c.buildKeywordIdeasRequest(seedKeywords, seedURL, opts)
	endpoint := fmt.Sprintf("%s/customers/%s:generateKeywordIdeas", c.baseURL, c.customerID)

	var raw generateKeywordIdeasResponse
	if err := c.post(ctx, endpoint, reqBody, &raw); err != nil {
		return nil, err
	}

	ideas := make([]KeywordIdea, 0, len(raw.Results))
	for _, r := range raw.Results {
		ideas = append(ideas, KeywordIdea{
			Text:                   r.Text,
			AvgMonthlySearches:     parseI64(r.KeywordIdeaMetrics.AvgMonthlySearches),
			Competition:            r.KeywordIdeaMetrics.Competition,
			LowTopOfPageBidMicros:  parseI64(r.KeywordIdeaMetrics.LowTopOfPageBidMicros),
			HighTopOfPageBidMicros: parseI64(r.KeywordIdeaMetrics.HighTopOfPageBidMicros),
		})
	}

	return &KeywordIdeasResponse{
		SeedKeywords: seedKeywords,
		URL:          seedURL,
		Ideas:        ideas,
		Count:        len(ideas),
	}, nil
}

// GetHistoricalMetrics returns historical search metrics for a list of keywords.
func (c *Client) GetHistoricalMetrics(
	ctx context.Context,
	keywords []string,
	opts HistoricalMetricsOptions,
) (*HistoricalMetricsResponse, error) {
	reqBody := generateHistoricalMetricsRequest{
		Keywords:             keywords,
		Language:             opts.Language,
		GeoTargetConstants:   opts.GeoTargetConstants,
		KeywordPlanNetwork:   opts.KeywordPlanNetwork,
		IncludeAdultKeywords: opts.IncludeAdultKeywords,
	}
	if len(opts.AggregateMetrics) > 0 {
		reqBody.AggregateMetrics = &aggregateMetrics{AggregateMetricTypes: opts.AggregateMetrics}
	}
	if opts.HistoricalDateRange != nil || opts.HistoricalIncludeAverageCpc {
		hmo := &historicalMetricsOptions{IncludeAverageCpc: opts.HistoricalIncludeAverageCpc}
		if opts.HistoricalDateRange != nil {
			hmo.YearMonthRange = &yearMonthRange{
				Start: yearMonth{Year: opts.HistoricalDateRange.Start.Year, Month: opts.HistoricalDateRange.Start.Month},
				End:   yearMonth{Year: opts.HistoricalDateRange.End.Year, Month: opts.HistoricalDateRange.End.Month},
			}
		}
		reqBody.HistoricalMetricsOptions = hmo
	}

	endpoint := fmt.Sprintf("%s/customers/%s:generateKeywordHistoricalMetrics", c.baseURL, c.customerID)

	var raw generateHistoricalMetricsResponse
	if err := c.post(ctx, endpoint, reqBody, &raw); err != nil {
		return nil, err
	}

	metrics := make([]KeywordMetrics, 0, len(raw.Results))
	for _, r := range raw.Results {
		monthly := make([]MonthlyVolume, 0, len(r.KeywordMetrics.MonthlySearchVolumes))
		for _, m := range r.KeywordMetrics.MonthlySearchVolumes {
			monthly = append(monthly, MonthlyVolume{
				Year:            int32(parseI64(m.Year)),
				Month:           parseMonthEnum(m.Month),
				MonthlySearches: parseI64(m.MonthlySearches),
			})
		}
		metrics = append(metrics, KeywordMetrics{
			Text:                   r.Text,
			AvgMonthlySearches:     parseI64(r.KeywordMetrics.AvgMonthlySearches),
			Competition:            r.KeywordMetrics.Competition,
			CompetitionIndex:       int32(parseI64(r.KeywordMetrics.CompetitionIndex)),
			LowTopOfPageBidMicros:  parseI64(r.KeywordMetrics.LowTopOfPageBidMicros),
			HighTopOfPageBidMicros: parseI64(r.KeywordMetrics.HighTopOfPageBidMicros),
			MonthlySearchVolumes:   monthly,
		})
	}

	return &HistoricalMetricsResponse{Keywords: metrics, Count: len(metrics)}, nil
}

// GetKeywordForecast returns projected performance metrics for a set of keywords.
func (c *Client) GetKeywordForecast(
	ctx context.Context,
	keywords []string,
	maxCPCMicros int64,
	forecastDays int,
	opts ForecastOptions,
) (*ForecastResponse, error) {
	if forecastDays <= 0 {
		forecastDays = 30
	}

	startDate := opts.StartDate
	endDate := opts.EndDate
	if startDate == "" && endDate == "" {
		now := time.Now().UTC()
		startDate = now.Format("2006-01-02")
		endDate = now.AddDate(0, 0, forecastDays).Format("2006-01-02")
	}

	defaultMatchType := opts.MatchType
	if defaultMatchType == "" {
		defaultMatchType = "BROAD"
	}

	keywordPlanNetwork := opts.KeywordPlanNetwork
	if keywordPlanNetwork == "" {
		keywordPlanNetwork = "GOOGLE_SEARCH"
	}

	specs := opts.KeywordSpecs
	if len(specs) == 0 {
		specs = make([]ForecastKeywordSpec, 0, len(keywords))
		for _, kw := range keywords {
			specs = append(specs, ForecastKeywordSpec{Text: kw})
		}
	}

	biddable := make([]adGroupForecastKeyword, 0, len(specs))
	for _, s := range specs {
		mt := s.MatchType
		if mt == "" {
			mt = defaultMatchType
		}
		entry := adGroupForecastKeyword{
			Keyword: forecastKeyword{Text: s.Text, MatchType: mt},
		}
		if s.MaxCpcBidMicros > 0 {
			entry.MaxCpcBidMicros = strconv.FormatInt(s.MaxCpcBidMicros, 10)
		}
		biddable = append(biddable, entry)
	}

	geoMods := make([]geoModifier, 0, len(opts.GeoTargetConstants))
	for _, g := range opts.GeoTargetConstants {
		geoMods = append(geoMods, geoModifier{GeoTargetConstant: g})
	}

	reqBody := generateForecastMetricsRequest{
		ForecastPeriod: forecastPeriod{
			StartDate: startDate,
			EndDate:   endDate,
		},
		Campaign: forecastCampaign{
			BiddingStrategy: campaignBiddingStrategy{
				ManualCpcBiddingStrategy: manualCpcBiddingStrategy{
					MaxCPCBidMicros: strconv.FormatInt(maxCPCMicros, 10),
				},
			},
			AdGroups:           []adGroupForecast{{Biddable: biddable}},
			GeoModifiers:       geoMods,
			LanguageConstants:  opts.LanguageConstants,
			KeywordPlanNetwork: keywordPlanNetwork,
		},
	}

	endpoint := fmt.Sprintf("%s/customers/%s:generateKeywordForecastMetrics", c.baseURL, c.customerID)

	var raw generateForecastMetricsResponse
	if err := c.post(ctx, endpoint, reqBody, &raw); err != nil {
		return nil, err
	}

	return &ForecastResponse{
		Campaign: CampaignForecastResult{
			Impressions:      raw.CampaignForecastMetrics.Impressions,
			ClickThroughRate: raw.CampaignForecastMetrics.ClickThroughRate,
			AverageCpcMicros: parseI64(raw.CampaignForecastMetrics.AverageCpcMicros),
			Clicks:           raw.CampaignForecastMetrics.Clicks,
			CostMicros:       parseI64(raw.CampaignForecastMetrics.CostMicros),
		},
		ForecastDays: forecastDays,
		MaxCPCMicros: maxCPCMicros,
	}, nil
}

func (c *Client) post(ctx context.Context, endpoint string, body, out any) error {
	reqBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshalling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("developer-token", c.developerToken)
	if c.loginCustomerID != "" {
		req.Header.Set("login-customer-id", c.loginCustomerID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Google Ads API returned HTTP %d: %s",
			resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return nil
}

func (c *Client) buildKeywordIdeasRequest(seedKeywords []string, seedURL string, opts KeywordIdeasOptions) generateKeywordIdeasRequest {
	req := generateKeywordIdeasRequest{
		Language:             opts.Language,
		GeoTargetConstants:   opts.GeoTargetConstants,
		KeywordPlanNetwork:   opts.KeywordPlanNetwork,
		IncludeAdultKeywords: opts.IncludeAdultKeywords,
		KeywordAnnotation:    opts.KeywordAnnotation,
		CurrencyCode:         opts.CurrencyCode,
		ToplevelDomain:       opts.ToplevelDomain,
	}
	if len(opts.AggregateMetrics) > 0 {
		req.AggregateMetrics = &aggregateMetrics{AggregateMetricTypes: opts.AggregateMetrics}
	}
	if opts.HistoricalDateRange != nil || opts.HistoricalIncludeAverageCpc {
		hmo := &historicalMetricsOptions{IncludeAverageCpc: opts.HistoricalIncludeAverageCpc}
		if opts.HistoricalDateRange != nil {
			hmo.YearMonthRange = &yearMonthRange{
				Start: yearMonth{Year: opts.HistoricalDateRange.Start.Year, Month: opts.HistoricalDateRange.Start.Month},
				End:   yearMonth{Year: opts.HistoricalDateRange.End.Year, Month: opts.HistoricalDateRange.End.Month},
			}
		}
		req.HistoricalMetricsOptions = hmo
	}
	switch {
	case len(seedKeywords) > 0 && seedURL != "":
		req.KeywordAndURLSeed = &keywordAndURLSeed{URL: seedURL, Keywords: seedKeywords}
	case seedURL != "":
		req.URLSeed = &urlSeed{URL: seedURL}
	default:
		req.KeywordSeed = &keywordSeed{Keywords: seedKeywords}
	}
	return req
}

func parseI64(s string) int64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// parseMonthEnum converts "JANUARY" → 1, etc.
func parseMonthEnum(month string) int32 {
	months := map[string]int32{
		"JANUARY": 1, "FEBRUARY": 2, "MARCH": 3, "APRIL": 4,
		"MAY": 5, "JUNE": 6, "JULY": 7, "AUGUST": 8,
		"SEPTEMBER": 9, "OCTOBER": 10, "NOVEMBER": 11, "DECEMBER": 12,
	}
	if v, ok := months[strings.ToUpper(month)]; ok {
		return v
	}
	return 0
}

// KeywordIdeasOptions are optional parameters for GenerateKeywordIdeas.
// All fields are optional; zero values are omitted from the request.
type KeywordIdeasOptions struct {
	Language                    string
	GeoTargetConstants          []string
	KeywordPlanNetwork          string // GOOGLE_SEARCH or GOOGLE_SEARCH_AND_PARTNERS
	IncludeAdultKeywords        bool
	KeywordAnnotation           []string
	AggregateMetrics            []string // currently only ["DEVICE"]
	HistoricalDateRange         *YearMonthRange
	HistoricalIncludeAverageCpc bool
	CurrencyCode                string
	ToplevelDomain              string
}

// YearMonthRange is exported so callers can build historical-metrics date ranges.
type YearMonthRange struct {
	Start YearMonth
	End   YearMonth
}

// YearMonth uses Google's enum month names ("JANUARY".."DECEMBER").
type YearMonth struct {
	Year  int32
	Month string
}

// HistoricalMetricsOptions are optional parameters for GetHistoricalMetrics.
type HistoricalMetricsOptions struct {
	Language                    string
	GeoTargetConstants          []string
	KeywordPlanNetwork          string
	IncludeAdultKeywords        bool
	AggregateMetrics            []string
	HistoricalDateRange         *YearMonthRange
	HistoricalIncludeAverageCpc bool
}

// ForecastOptions are optional parameters for GetKeywordForecast.
type ForecastOptions struct {
	GeoTargetConstants []string
	LanguageConstants  []string
	KeywordPlanNetwork string
	MatchType          string // EXACT, PHRASE, or BROAD (default BROAD)
	StartDate          string // YYYY-MM-DD; if set, both StartDate and EndDate must be set
	EndDate            string
	// KeywordSpecs is the per-keyword shape — overrides the keywords arg when len > 0.
	KeywordSpecs []ForecastKeywordSpec
}

// ForecastKeywordSpec is one keyword with optional per-keyword overrides.
// MatchType defaults to ForecastOptions.MatchType (or BROAD).
// MaxCpcBidMicros, when > 0, is sent as a per-keyword bid override.
type ForecastKeywordSpec struct {
	Text            string
	MatchType       string
	MaxCpcBidMicros int64
}

