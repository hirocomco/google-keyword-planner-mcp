package keywordplanner_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ncosentino/google-keyword-planner-mcp/go/internal/keywordplanner"
)

// captureRequestBody runs an httptest server that records the request body bytes
// and returns a stub response. Returns the captured body after the test call.
func captureRequestBody(t *testing.T, response any) (*httptest.Server, *[]byte) {
	t.Helper()
	captured := new([]byte)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*captured = body
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	return srv, captured
}

func TestGenerateKeywordIdeas_AllOptionsReachWire(t *testing.T) {
	t.Parallel()
	srv, captured := captureRequestBody(t, map[string]any{"results": []any{}})
	defer srv.Close()

	client := keywordplanner.NewTestClient("dev", "123", "", srv.URL, srv.Client())

	opts := keywordplanner.KeywordIdeasOptions{
		Language:             "languageConstants/1000",
		GeoTargetConstants:   []string{"geoTargetConstants/21137"},
		KeywordPlanNetwork:   "GOOGLE_SEARCH",
		IncludeAdultKeywords: true,
		KeywordAnnotation:    []string{"KEYWORD_CONCEPT"},
		AggregateMetrics:     []string{"DEVICE"},
		HistoricalDateRange: &keywordplanner.YearMonthRange{
			Start: keywordplanner.YearMonth{Year: 2024, Month: "JANUARY"},
			End:   keywordplanner.YearMonth{Year: 2024, Month: "DECEMBER"},
		},
		HistoricalIncludeAverageCpc: true,
		CurrencyCode:                "USD",
		ToplevelDomain:              "com",
	}

	_, _ = client.GenerateKeywordIdeas(context.Background(), []string{"lawn care"}, "", opts)

	var got map[string]any
	if err := json.Unmarshal(*captured, &got); err != nil {
		t.Fatalf("captured body not JSON: %v\n%s", err, string(*captured))
	}

	if got["language"] != "languageConstants/1000" {
		t.Errorf("language = %v", got["language"])
	}
	geos, _ := got["geoTargetConstants"].([]any)
	if len(geos) != 1 || geos[0] != "geoTargetConstants/21137" {
		t.Errorf("geoTargetConstants = %v", got["geoTargetConstants"])
	}
	if got["keywordPlanNetwork"] != "GOOGLE_SEARCH" {
		t.Errorf("keywordPlanNetwork = %v", got["keywordPlanNetwork"])
	}
	if got["includeAdultKeywords"] != true {
		t.Errorf("includeAdultKeywords = %v", got["includeAdultKeywords"])
	}
	if got["currencyCode"] != "USD" {
		t.Errorf("currencyCode = %v", got["currencyCode"])
	}
	if got["topLevelDomain"] != "com" {
		t.Errorf("topLevelDomain = %v", got["topLevelDomain"])
	}
	hmo, _ := got["historicalMetricsOptions"].(map[string]any)
	if hmo == nil || hmo["includeAverageCpc"] != true {
		t.Errorf("historicalMetricsOptions = %v", got["historicalMetricsOptions"])
	}
	am, _ := got["aggregateMetrics"].(map[string]any)
	if am == nil {
		t.Errorf("aggregateMetrics missing")
	}

	anns, _ := got["keywordAnnotation"].([]any)
	if len(anns) != 1 || anns[0] != "KEYWORD_CONCEPT" {
		t.Errorf("keywordAnnotation = %v", got["keywordAnnotation"])
	}

	ymr, _ := hmo["yearMonthRange"].(map[string]any)
	if ymr == nil {
		t.Errorf("yearMonthRange missing")
	} else {
		start, _ := ymr["start"].(map[string]any)
		end, _ := ymr["end"].(map[string]any)
		if start == nil || end == nil {
			t.Errorf("yearMonthRange start/end missing: start=%v end=%v", start, end)
		} else {
			if v, _ := start["year"].(float64); int(v) != 2024 || start["month"] != "JANUARY" {
				t.Errorf("yearMonthRange.start = %v", start)
			}
			if v, _ := end["year"].(float64); int(v) != 2024 || end["month"] != "DECEMBER" {
				t.Errorf("yearMonthRange.end = %v", end)
			}
		}
	}
}

func TestGenerateKeywordIdeas_OmittedOptions_NoNewFields(t *testing.T) {
	t.Parallel()
	srv, captured := captureRequestBody(t, map[string]any{"results": []any{}})
	defer srv.Close()

	client := keywordplanner.NewTestClient("dev", "123", "", srv.URL, srv.Client())
	_, _ = client.GenerateKeywordIdeas(context.Background(), []string{"lawn care"}, "", keywordplanner.KeywordIdeasOptions{})

	var got map[string]any
	if err := json.Unmarshal(*captured, &got); err != nil {
		t.Fatalf("captured body not JSON: %v\n%s", err, string(*captured))
	}

	for _, key := range []string{
		"geoTargetConstants", "keywordPlanNetwork", "includeAdultKeywords",
		"keywordAnnotation", "aggregateMetrics", "historicalMetricsOptions",
		"currencyCode", "topLevelDomain",
	} {
		if _, present := got[key]; present {
			t.Errorf("expected %q to be omitted from request body, got %v", key, got[key])
		}
	}
}

func TestGetHistoricalMetrics_AllOptionsReachWire(t *testing.T) {
	t.Parallel()
	srv, captured := captureRequestBody(t, map[string]any{"metrics": []any{}})
	defer srv.Close()

	client := keywordplanner.NewTestClient("dev", "123", "", srv.URL, srv.Client())

	opts := keywordplanner.HistoricalMetricsOptions{
		Language:             "languageConstants/1000",
		GeoTargetConstants:   []string{"geoTargetConstants/21137"},
		KeywordPlanNetwork:   "GOOGLE_SEARCH",
		IncludeAdultKeywords: true,
		AggregateMetrics:     []string{"DEVICE"},
		HistoricalDateRange: &keywordplanner.YearMonthRange{
			Start: keywordplanner.YearMonth{Year: 2024, Month: "JANUARY"},
			End:   keywordplanner.YearMonth{Year: 2024, Month: "DECEMBER"},
		},
		HistoricalIncludeAverageCpc: true,
	}

	_, _ = client.GetHistoricalMetrics(context.Background(), []string{"lawn care"}, opts)

	var got map[string]any
	if err := json.Unmarshal(*captured, &got); err != nil {
		t.Fatalf("captured body not JSON: %v", err)
	}

	if got["language"] != "languageConstants/1000" {
		t.Errorf("language = %v", got["language"])
	}
	geos, _ := got["geoTargetConstants"].([]any)
	if len(geos) != 1 || geos[0] != "geoTargetConstants/21137" {
		t.Errorf("geoTargetConstants = %v", got["geoTargetConstants"])
	}
	if got["keywordPlanNetwork"] != "GOOGLE_SEARCH" {
		t.Errorf("keywordPlanNetwork = %v", got["keywordPlanNetwork"])
	}
	if got["includeAdultKeywords"] != true {
		t.Errorf("includeAdultKeywords = %v", got["includeAdultKeywords"])
	}
	hmo, _ := got["historicalMetricsOptions"].(map[string]any)
	if hmo == nil {
		t.Fatalf("historicalMetricsOptions missing")
	}
	if hmo["includeAverageCpc"] != true {
		t.Errorf("includeAverageCpc = %v", hmo["includeAverageCpc"])
	}
	ymr, _ := hmo["yearMonthRange"].(map[string]any)
	if ymr == nil {
		t.Fatalf("yearMonthRange missing")
	}
	start, _ := ymr["start"].(map[string]any)
	end, _ := ymr["end"].(map[string]any)
	if start == nil || end == nil {
		t.Fatalf("yearMonthRange start/end missing: start=%v end=%v", start, end)
	}
	if v, _ := start["year"].(float64); int(v) != 2024 || start["month"] != "JANUARY" {
		t.Errorf("yearMonthRange.start = %v", start)
	}
	if v, _ := end["year"].(float64); int(v) != 2024 || end["month"] != "DECEMBER" {
		t.Errorf("yearMonthRange.end = %v", end)
	}
	am, _ := got["aggregateMetrics"].(map[string]any)
	if am == nil {
		t.Errorf("aggregateMetrics missing")
	} else {
		types, _ := am["aggregateMetricTypes"].([]any)
		if len(types) != 1 || types[0] != "DEVICE" {
			t.Errorf("aggregateMetricTypes = %v", types)
		}
	}
}

func TestGetHistoricalMetrics_OmittedOptions_NoNewFields(t *testing.T) {
	t.Parallel()
	srv, captured := captureRequestBody(t, map[string]any{"metrics": []any{}})
	defer srv.Close()

	client := keywordplanner.NewTestClient("dev", "123", "", srv.URL, srv.Client())
	_, _ = client.GetHistoricalMetrics(context.Background(), []string{"lawn care"}, keywordplanner.HistoricalMetricsOptions{})

	var got map[string]any
	if err := json.Unmarshal(*captured, &got); err != nil {
		t.Fatalf("captured body not JSON: %v\n%s", err, string(*captured))
	}

	for _, key := range []string{
		"language", "geoTargetConstants", "keywordPlanNetwork",
		"includeAdultKeywords", "aggregateMetrics", "historicalMetricsOptions",
	} {
		if _, present := got[key]; present {
			t.Errorf("expected %q to be omitted, got %v", key, got[key])
		}
	}
}

func TestGetKeywordForecast_AllBaseOptionsReachWire(t *testing.T) {
	t.Parallel()
	srv, captured := captureRequestBody(t, map[string]any{"adGroupForecastMetrics": []any{}})
	defer srv.Close()

	client := keywordplanner.NewTestClient("dev", "123", "", srv.URL, srv.Client())

	opts := keywordplanner.ForecastOptions{
		GeoTargetConstants: []string{"geoTargetConstants/21137"},
		LanguageConstants:  []string{"languageConstants/1000"},
		KeywordPlanNetwork: "GOOGLE_SEARCH",
		MatchType:          "EXACT",
		StartDate:          "2026-05-01",
		EndDate:            "2026-05-30",
	}

	_, _ = client.GetKeywordForecast(context.Background(),
		[]string{"lawn care", "fertilizer"}, 2_000_000, 30, opts)

	var got map[string]any
	if err := json.Unmarshal(*captured, &got); err != nil {
		t.Fatalf("captured body not JSON: %v\n%s", err, string(*captured))
	}
	spec, _ := got["campaignForecastSpec"].(map[string]any)
	if spec == nil {
		t.Fatalf("missing campaignForecastSpec: %s", string(*captured))
	}
	if spec["startDate"] != "2026-05-01" || spec["endDate"] != "2026-05-30" {
		t.Errorf("dates = %v / %v", spec["startDate"], spec["endDate"])
	}
	if spec["keywordPlanNetwork"] != "GOOGLE_SEARCH" {
		t.Errorf("keywordPlanNetwork = %v", spec["keywordPlanNetwork"])
	}
	langs, _ := spec["languageConstants"].([]any)
	if len(langs) != 1 || langs[0] != "languageConstants/1000" {
		t.Errorf("languageConstants = %v", spec["languageConstants"])
	}
	geos, _ := spec["geoModifiers"].([]any)
	if len(geos) != 1 {
		t.Fatalf("geoModifiers = %v", spec["geoModifiers"])
	}
	g0, _ := geos[0].(map[string]any)
	if g0["geoTargetConstant"] != "geoTargetConstants/21137" {
		t.Errorf("geoModifiers[0].geoTargetConstant = %v", g0["geoTargetConstant"])
	}
	ags, _ := spec["adGroups"].([]any)
	if len(ags) != 1 {
		t.Fatalf("adGroups = %v", ags)
	}
	bk, _ := ags[0].(map[string]any)["biddableKeywords"].([]any)
	if len(bk) != 2 {
		t.Fatalf("biddableKeywords len = %d", len(bk))
	}
	for i, b := range bk {
		bm, _ := b.(map[string]any)
		kw, _ := bm["keyword"].(map[string]any)
		if kw["matchType"] != "EXACT" {
			t.Errorf("keyword[%d].matchType = %v", i, kw["matchType"])
		}
	}
}

func TestGetKeywordForecast_DefaultMatchTypeBroad(t *testing.T) {
	t.Parallel()
	srv, captured := captureRequestBody(t, map[string]any{"adGroupForecastMetrics": []any{}})
	defer srv.Close()

	client := keywordplanner.NewTestClient("dev", "123", "", srv.URL, srv.Client())
	_, _ = client.GetKeywordForecast(context.Background(),
		[]string{"k"}, 1_000_000, 30, keywordplanner.ForecastOptions{})

	var got map[string]any
	if err := json.Unmarshal(*captured, &got); err != nil {
		t.Fatalf("captured body not JSON: %v\n%s", err, string(*captured))
	}
	spec, _ := got["campaignForecastSpec"].(map[string]any)
	if spec == nil {
		t.Fatalf("missing campaignForecastSpec")
	}
	ags, _ := spec["adGroups"].([]any)
	bk, _ := ags[0].(map[string]any)["biddableKeywords"].([]any)
	kw, _ := bk[0].(map[string]any)["keyword"].(map[string]any)
	if kw["matchType"] != "BROAD" {
		t.Errorf("default matchType = %v, want BROAD", kw["matchType"])
	}
}
