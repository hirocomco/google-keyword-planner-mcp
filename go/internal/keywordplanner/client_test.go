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
