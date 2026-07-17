package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"techmind/internal/pkg/settings"
)

// PrometheusSnapshot 查询与告警类型匹配的聚合指标，控制返回量以适合 LLM 诊断输入。
func PrometheusSnapshot(ctx context.Context, alertName string) Evidence {
	baseURL := strings.TrimRight(settings.Conf.Monitor.PrometheusURL, "/")
	if baseURL == "" {
		return Evidence{"prometheus_status": "not configured"}
	}

	queries := map[string]string{
		"api_request_rate": "sum(rate(http_requests_total{path!=\"/metrics\"}[5m]))",
		"api_error_rate":   "sum(rate(http_errors_total[5m])) / clamp_min(sum(rate(http_requests_total{path!=\"/metrics\"}[5m])), 1)",
	}
	switch alertName {
	case "SearchLatencyHigh":
		queries["search_p95_seconds"] = "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path=\"/api/v1/search\"}[5m])) by (le))"
	case "APILatencyHigh":
		queries["api_p95_seconds"] = "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path!=\"/metrics\"}[5m])) by (le))"
	case "SlowRequestSpike":
		queries["slow_request_rate"] = "sum(rate(slow_requests_total[5m]))"
	case "RedisStreamBacklogHigh":
		queries["redis_stream_pending"] = "sum(redis_stream_pending_total)"
	case "WorkerConsumeLagHigh":
		queries["worker_consume_p95_seconds"] = "histogram_quantile(0.95, sum(rate(redis_stream_consume_duration_seconds_bucket[5m])) by (le))"
	case "CacheHitRateLow":
		queries["cache_hit_rate"] = "sum(rate(cache_hit_total[5m])) / clamp_min(sum(rate(cache_hit_total[5m])) + sum(rate(cache_miss_total[5m])), 1)"
	case "AICallFailureHigh":
		queries["ai_failure_rate"] = "sum(rate(ai_call_errors_total[5m])) / clamp_min(sum(rate(ai_calls_total[5m])), 1)"
	case "AICallLatencyHigh":
		queries["ai_p95_seconds"] = "histogram_quantile(0.95, sum(rate(ai_call_duration_seconds_bucket[5m])) by (le))"
	case "MilvusSearchLatencyHigh":
		queries["milvus_p95_seconds"] = "histogram_quantile(0.95, sum(rate(milvus_search_duration_seconds_bucket[5m])) by (le))"
	case "MilvusSearchErrorHigh":
		queries["milvus_error_rate"] = "sum(rate(milvus_search_errors_total[5m]))"
	case "OpsDiagnoseDurationHigh":
		queries["ops_p95_seconds"] = "histogram_quantile(0.95, sum(rate(ops_diagnose_duration_seconds_bucket[5m])) by (le))"
	case "PodRestartHigh":
		queries["pod_restart_increase"] = "sum(increase(kube_pod_container_status_restarts_total[5m]))"
	}

	evidence := Evidence{}
	for name, query := range queries {
		result, err := queryPrometheus(ctx, baseURL, query)
		if err != nil {
			evidence["prometheus_"+name+"_error"] = err.Error()
			continue
		}
		evidence["prometheus_"+name] = result
	}
	return evidence
}

// PrometheusRangeSnapshot 查询告警对应指标在受限时间窗内的趋势。
func PrometheusRangeSnapshot(ctx context.Context, alertName string, start, end time.Time) Evidence {
	baseURL := strings.TrimRight(settings.Conf.Monitor.PrometheusURL, "/")
	if baseURL == "" {
		return Evidence{"prometheus_range_status": "not configured"}
	}
	if end.IsZero() {
		end = time.Now()
	}
	if start.IsZero() || !end.After(start) {
		start = end.Add(-30 * time.Minute)
	}
	if end.Sub(start) > time.Hour {
		start = end.Add(-time.Hour)
	}
	metricName, query := prometheusRangeQueryForAlert(alertName)
	step := end.Sub(start) / 60
	if step < 15*time.Second {
		step = 15 * time.Second
	}
	if step > time.Minute {
		step = time.Minute
	}
	result, err := queryPrometheusRange(ctx, baseURL, query, start, end, step)
	if err != nil {
		return Evidence{"prometheus_range_error": err.Error()}
	}
	return Evidence{
		"prometheus_range_metric": metricName,
		"prometheus_range_start":  start.UTC().Format(time.RFC3339),
		"prometheus_range_end":    end.UTC().Format(time.RFC3339),
		"prometheus_range_values": result,
	}
}

func prometheusRangeQueryForAlert(alertName string) (string, string) {
	switch alertName {
	case "SearchLatencyHigh":
		return "search_p95_seconds", "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path=\"/api/v1/search\"}[5m])) by (le))"
	case "APILatencyHigh":
		return "api_p95_seconds", "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path!=\"/metrics\"}[5m])) by (le))"
	case "APIHighErrorRate":
		return "api_error_rate", "sum(rate(http_errors_total[5m])) / clamp_min(sum(rate(http_requests_total{path!=\"/metrics\"}[5m])), 1)"
	case "SlowRequestSpike":
		return "slow_request_rate", "sum(rate(slow_requests_total[5m]))"
	case "RedisStreamBacklogHigh":
		return "redis_stream_pending", "sum(redis_stream_pending_total)"
	case "WorkerConsumeLagHigh":
		return "worker_consume_p95_seconds", "histogram_quantile(0.95, sum(rate(redis_stream_consume_duration_seconds_bucket[5m])) by (le))"
	case "CacheHitRateLow":
		return "cache_hit_rate", "sum(rate(cache_hit_total[5m])) / clamp_min(sum(rate(cache_hit_total[5m])) + sum(rate(cache_miss_total[5m])), 1)"
	case "AICallFailureHigh":
		return "ai_failure_rate", "sum(rate(ai_call_errors_total[5m])) / clamp_min(sum(rate(ai_calls_total[5m])), 1)"
	case "AICallLatencyHigh":
		return "ai_p95_seconds", "histogram_quantile(0.95, sum(rate(ai_call_duration_seconds_bucket[5m])) by (le))"
	case "MilvusSearchLatencyHigh":
		return "milvus_p95_seconds", "histogram_quantile(0.95, sum(rate(milvus_search_duration_seconds_bucket[5m])) by (le))"
	case "MilvusSearchErrorHigh":
		return "milvus_error_rate", "sum(rate(milvus_search_errors_total[5m]))"
	case "OpsDiagnoseDurationHigh":
		return "ops_p95_seconds", "histogram_quantile(0.95, sum(rate(ops_diagnose_duration_seconds_bucket[5m])) by (le))"
	case "PodRestartHigh":
		return "pod_restart_increase", "sum(increase(kube_pod_container_status_restarts_total[5m]))"
	default:
		return "api_request_rate", "sum(rate(http_requests_total{path!=\"/metrics\"}[5m]))"
	}
}

func queryPrometheus(ctx context.Context, baseURL, query string) (string, error) {
	endpoint := baseURL + "/api/v1/query?query=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("prometheus returned %s", resp.Status)
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Result json.RawMessage `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Status != "success" {
		return "", fmt.Errorf("prometheus query status %q", payload.Status)
	}
	result := string(payload.Data.Result)
	if len(result) > 2000 {
		result = result[:2000] + "..."
	}
	return result, nil
}

func queryPrometheusRange(ctx context.Context, baseURL, query string, start, end time.Time, step time.Duration) (string, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", start.Unix()))
	params.Set("end", fmt.Sprintf("%d", end.Unix()))
	params.Set("step", fmt.Sprintf("%ds", int(step.Seconds())))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/query_range?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("prometheus returned %s", resp.Status)
	}
	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Result json.RawMessage `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Status != "success" {
		return "", fmt.Errorf("prometheus query status %q", payload.Status)
	}
	result := string(payload.Data.Result)
	if len(result) > 4000 {
		result = result[:4000] + "..."
	}
	return result, nil
}
