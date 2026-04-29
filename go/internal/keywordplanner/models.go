// Package keywordplanner provides types for the Google Ads Keyword Planner API.
package keywordplanner

// KeywordIdea is a keyword suggestion with historical performance metrics.
type KeywordIdea struct {
	Text        string  `json:"text"`
	AvgMonthlySearches int64  `json:"avgMonthlySearches"`
	Competition string  `json:"competition"`
	LowTopOfPageBidMicros  int64 `json:"lowTopOfPageBidMicros,omitempty"`
	HighTopOfPageBidMicros int64 `json:"highTopOfPageBidMicros,omitempty"`
}

// KeywordIdeasResponse is the result of generating keyword ideas.
type KeywordIdeasResponse struct {
	SeedKeywords []string      `json:"seedKeywords,omitempty"`
	URL          string        `json:"url,omitempty"`
	Ideas        []KeywordIdea `json:"ideas"`
	Count        int           `json:"count"`
}

// KeywordMetrics holds historical search metrics for a single keyword.
type KeywordMetrics struct {
	Text               string        `json:"text"`
	AvgMonthlySearches int64         `json:"avgMonthlySearches"`
	Competition        string        `json:"competition"`
	CompetitionIndex   int32         `json:"competitionIndex"`
	LowTopOfPageBidMicros  int64     `json:"lowTopOfPageBidMicros,omitempty"`
	HighTopOfPageBidMicros int64     `json:"highTopOfPageBidMicros,omitempty"`
	MonthlySearchVolumes []MonthlyVolume `json:"monthlySearchVolumes,omitempty"`
}

// MonthlyVolume is the search volume for a specific month.
type MonthlyVolume struct {
	Year  int32 `json:"year"`
	Month int32 `json:"month"`
	MonthlySearches int64 `json:"monthlySearches"`
}

// HistoricalMetricsResponse is the result of a historical metrics lookup.
type HistoricalMetricsResponse struct {
	Keywords []KeywordMetrics `json:"keywords"`
	Count    int              `json:"count"`
}

// KeywordForecastMetrics holds projected performance for a keyword.
type KeywordForecastMetrics struct {
	Text        string  `json:"text"`
	Impressions float64 `json:"impressions"`
	Clicks      float64 `json:"clicks"`
	CostMicros  float64 `json:"costMicros"`
	CTR         float64 `json:"ctr"`
}

// ForecastResponse is the result of a keyword forecast request.
type ForecastResponse struct {
	Keywords  []KeywordForecastMetrics `json:"keywords"`
	ForecastDays int                  `json:"forecastDays"`
	MaxCPCMicros int64                `json:"maxCpcMicros"`
}

// --- Google Ads API raw request/response types ---

type generateKeywordIdeasRequest struct {
	CustomerID               string                    `json:"customerId,omitempty"`
	Language                 string                    `json:"language,omitempty"`
	GeoTargetConstants       []string                  `json:"geoTargetConstants,omitempty"`
	KeywordSeed              *keywordSeed              `json:"keywordSeed,omitempty"`
	URLSeed                  *urlSeed                  `json:"urlSeed,omitempty"`
	KeywordAndURLSeed        *keywordAndURLSeed        `json:"keywordAndUrlSeed,omitempty"`
	KeywordPlanNetwork       string                    `json:"keywordPlanNetwork,omitempty"`
	IncludeAdultKeywords     bool                      `json:"includeAdultKeywords,omitempty"`
	KeywordAnnotation        []string                  `json:"keywordAnnotation,omitempty"`
	AggregateMetrics         *aggregateMetrics         `json:"aggregateMetrics,omitempty"`
	HistoricalMetricsOptions *historicalMetricsOptions `json:"historicalMetricsOptions,omitempty"`
	CurrencyCode             string                    `json:"currencyCode,omitempty"`
	ToplevelDomain           string                    `json:"topLevelDomain,omitempty"`
}

type keywordSeed struct {
	Keywords []string `json:"keywords"`
}

type urlSeed struct {
	URL string `json:"url"`
}

type keywordAndURLSeed struct {
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}

type generateKeywordIdeasResponse struct {
	Results []keywordIdeaResult `json:"results"`
}

type keywordIdeaResult struct {
	Text            string              `json:"text"`
	KeywordIdeaMetrics keywordIdeaMetrics `json:"keywordIdeaMetrics"`
}

type keywordIdeaMetrics struct {
	AvgMonthlySearches     string `json:"avgMonthlySearches"`
	Competition            string `json:"competition"`
	CompetitionIndex       string `json:"competitionIndex"`
	LowTopOfPageBidMicros  string `json:"lowTopOfPageBidMicros"`
	HighTopOfPageBidMicros string `json:"highTopOfPageBidMicros"`
}

type generateHistoricalMetricsRequest struct {
	Keywords                 []string                  `json:"keywords"`
	Language                 string                    `json:"language,omitempty"`
	GeoTargetConstants       []string                  `json:"geoTargetConstants,omitempty"`
	KeywordPlanNetwork       string                    `json:"keywordPlanNetwork,omitempty"`
	IncludeAdultKeywords     bool                      `json:"includeAdultKeywords,omitempty"`
	AggregateMetrics         *aggregateMetrics         `json:"aggregateMetrics,omitempty"`
	HistoricalMetricsOptions *historicalMetricsOptions `json:"historicalMetricsOptions,omitempty"`
}

type generateHistoricalMetricsResponse struct {
	Metrics []historicalMetricsResult `json:"metrics"`
}

type historicalMetricsResult struct {
	Text           string           `json:"text"`
	KeywordMetrics historicalMetrics `json:"keywordMetrics"`
}

type historicalMetrics struct {
	AvgMonthlySearches     string                 `json:"avgMonthlySearches"`
	Competition            string                 `json:"competition"`
	CompetitionIndex       int32                  `json:"competitionIndex"`
	LowTopOfPageBidMicros  string                 `json:"lowTopOfPageBidMicros"`
	HighTopOfPageBidMicros string                 `json:"highTopOfPageBidMicros"`
	MonthlySearchVolumes   []monthlySearchVolume  `json:"monthlySearchVolumes"`
}

type monthlySearchVolume struct {
	Year            int32  `json:"year"`
	Month           string `json:"month"`
	MonthlySearches string `json:"monthlySearches"`
}

type generateForecastMetricsRequest struct {
	CampaignForecastSpec campaignForecastSpec `json:"campaignForecastSpec"`
}

type campaignForecastSpec struct {
	BiddingStrategy    biddingStrategy   `json:"biddingStrategy"`
	StartDate          string            `json:"startDate"`
	EndDate            string            `json:"endDate"`
	AdGroups           []adGroupForecast `json:"adGroups"`
	GeoModifiers       []geoModifier     `json:"geoModifiers,omitempty"`
	LanguageConstants  []string          `json:"languageConstants,omitempty"`
	KeywordPlanNetwork string            `json:"keywordPlanNetwork,omitempty"`
}

type biddingStrategy struct {
	ManualCpcBiddingStrategy manualCpcBiddingStrategy `json:"manualCpcBiddingStrategy"`
}

type manualCpcBiddingStrategy struct {
	MaxCPCBidMicros string `json:"maxCpcBidMicros"`
}

type adGroupForecast struct {
	Biddable []adGroupForecastKeyword `json:"biddableKeywords"`
}

type adGroupForecastKeyword struct {
	Keyword         forecastKeyword `json:"keyword"`
	MaxCpcBidMicros string          `json:"maxCpcBidMicros,omitempty"`
}

type forecastKeyword struct {
	Text      string `json:"text"`
	MatchType string `json:"matchType"`
}

type generateForecastMetricsResponse struct {
	AdGroupForecastMetrics []adGroupForecastMetrics `json:"adGroupForecastMetrics"`
}

type adGroupForecastMetrics struct {
	KeywordForecastMetrics []keywordForecastMetric `json:"keywordForecastMetrics"`
}

type keywordForecastMetric struct {
	Keyword  forecastKeyword    `json:"keyword"`
	Metrics  forecastMetricData `json:"metrics"`
}

type forecastMetricData struct {
	Impressions float64 `json:"impressions"`
	Clicks      float64 `json:"clicks"`
	CostMicros  float64 `json:"costMicros"`
	CTR         float64 `json:"ctr"`
}

// historicalMetricsOptions controls historical metrics data range and breakdown.
type historicalMetricsOptions struct {
	YearMonthRange    *yearMonthRange `json:"yearMonthRange,omitempty"`
	IncludeAverageCpc bool            `json:"includeAverageCpc,omitempty"`
}

type yearMonthRange struct {
	Start yearMonth `json:"start"`
	End   yearMonth `json:"end"`
}

type yearMonth struct {
	Year  int32  `json:"year"`
	Month string `json:"month"` // "JANUARY".."DECEMBER"
}

// aggregateMetrics requests aggregate breakdowns (currently only DEVICE supported by Google).
type aggregateMetrics struct {
	AggregateMetricTypes []string `json:"aggregateMetricTypes,omitempty"`
}

// geoModifier maps a single geo target inside campaignForecastSpec.geoModifiers.
type geoModifier struct {
	GeoTargetConstant string `json:"geoTargetConstant"`
}
