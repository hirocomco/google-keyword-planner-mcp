# google-keyword-planner-mcp (hirocomco fork)

Hirocom fork of [`ncosentino/google-keyword-planner-mcp`](https://github.com/ncosentino/google-keyword-planner-mcp)
v0.1.0 (MIT). Adds the missing `geoTargetConstants`, `keywordPlanNetwork`,
language, date-range, match-type, and per-keyword bid parameters to all
three tools. Upstream ships none of these on the wire.

## Tools (parameters new in this fork are marked NEW)

### `generate_keyword_ideas`

| Parameter | Notes |
|---|---|
| `seed_keywords`, `url` | Same as upstream. |
| `language` | Same as upstream. |
| `geo_target_constants` | NEW. e.g. `["geoTargetConstants/21137"]` (California). |
| `keyword_plan_network` | NEW. `GOOGLE_SEARCH` or `GOOGLE_SEARCH_AND_PARTNERS`. |
| `include_adult_keywords` | NEW. |
| `keyword_annotation` | NEW. e.g. `["KEYWORD_CONCEPT"]`. |
| `aggregate_metrics` | NEW. Currently only `["DEVICE"]`. |
| `historical_metrics_start`, `historical_metrics_end` | NEW. `{year, month}` enum. |
| `include_average_cpc` | NEW. |
| `currency_code` | NEW. ISO 4217 (`USD`, `EUR`, ...). |
| `toplevel_domain` | NEW. `com`, `co.uk`, ... |

### `get_historical_metrics`

| Parameter | Notes |
|---|---|
| `keywords` | Same as upstream. |
| `language` | NEW. |
| `geo_target_constants`, `keyword_plan_network`, `include_adult_keywords`, `aggregate_metrics`, `historical_metrics_start/end`, `include_average_cpc` | NEW. |

### `get_keyword_forecast`

| Parameter | Notes |
|---|---|
| `keywords`, `max_cpc_micros`, `forecast_days` | Same as upstream (forecast_days is overridden when explicit dates are set). |
| `keyword_specs` | NEW. Per-keyword shape: `[{text, match_type?, max_cpc_micros?}]`. Mutually exclusive with `keywords`. |
| `geo_target_constants`, `language_constants`, `keyword_plan_network` | NEW. |
| `match_type` | NEW. Global default for keywords (`EXACT`/`PHRASE`/`BROAD`). |
| `start_date`, `end_date` | NEW. `YYYY-MM-DD`. Both or neither. |

## Build

```bash
cd go
go build ./...
go test ./...
```

## Releases

GitHub Actions builds `darwin-amd64`, `darwin-arm64`, and `linux-amd64`
binaries on tag push. Tags follow `v0.1.0-hirocom.N`.

## License

MIT (inherited from upstream).
