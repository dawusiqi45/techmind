package monitor

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests."},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "HTTP request duration in seconds.", Buckets: prometheus.DefBuckets},
		[]string{"method", "path"},
	)
	httpErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_errors_total", Help: "Total HTTP 5xx errors."},
		[]string{"method", "path", "status"},
	)
	slowRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "slow_requests_total", Help: "Total slow HTTP requests."},
		[]string{"method", "path"},
	)
	monitorErrorEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "monitor_error_events_total", Help: "Total monitor error events."},
		[]string{"source"},
	)
	redisStreamPendingTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "redis_stream_pending_total", Help: "Redis Stream pending messages."},
		[]string{"stream", "group"},
	)
	redisStreamLenTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "redis_stream_len_total", Help: "Redis Stream length."},
		[]string{"stream"},
	)
	articleSearchDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "article_search_duration_seconds", Help: "Article search duration in seconds.", Buckets: prometheus.DefBuckets},
		[]string{"stage"},
	)
	milvusSearchDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{Name: "milvus_search_duration_seconds", Help: "Milvus search duration in seconds.", Buckets: prometheus.DefBuckets},
	)
	milvusSearchErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "milvus_search_errors_total", Help: "Total Milvus search errors."},
	)
	aiCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ai_calls_total", Help: "Total AI calls."},
		[]string{"kind", "status"},
	)
	aiCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "ai_call_duration_seconds", Help: "AI call duration in seconds.", Buckets: prometheus.DefBuckets},
		[]string{"kind"},
	)
	aiCallErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ai_call_errors_total", Help: "Total AI call errors."},
		[]string{"kind"},
	)
	workerTasksTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "worker_tasks_total", Help: "Worker task results."},
		[]string{"task_type", "status"},
	)
)

// RegisterMetrics 注册项目自定义 Prometheus 指标。重复注册时忽略 AlreadyRegistered 错误。
func RegisterMetrics() {
	collectors := []prometheus.Collector{
		httpRequestsTotal,
		httpRequestDuration,
		httpErrorsTotal,
		slowRequestsTotal,
		monitorErrorEventsTotal,
		redisStreamPendingTotal,
		redisStreamLenTotal,
		articleSearchDuration,
		milvusSearchDuration,
		milvusSearchErrorsTotal,
		aiCallsTotal,
		aiCallDuration,
		aiCallErrorsTotal,
		workerTasksTotal,
	}
	for _, c := range collectors {
		if err := prometheus.Register(c); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				continue
			}
		}
	}
}

func ObserveHTTPRequest(method, path string, status int, duration time.Duration) {
	statusText := strconv.Itoa(status)
	httpRequestsTotal.WithLabelValues(method, path, statusText).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	if status >= 500 {
		httpErrorsTotal.WithLabelValues(method, path, statusText).Inc()
	}
}

func IncSlowRequest(method, path string) {
	slowRequestsTotal.WithLabelValues(method, path).Inc()
}

func IncErrorEvent(source string) {
	monitorErrorEventsTotal.WithLabelValues(source).Inc()
}

func SetRedisStreamPending(stream, group string, count int64) {
	redisStreamPendingTotal.WithLabelValues(stream, group).Set(float64(count))
}

func SetRedisStreamLen(stream string, count int64) {
	redisStreamLenTotal.WithLabelValues(stream).Set(float64(count))
}

func ObserveArticleSearch(stage string, duration time.Duration) {
	articleSearchDuration.WithLabelValues(stage).Observe(duration.Seconds())
}

func ObserveMilvusSearch(duration time.Duration, err error) {
	milvusSearchDuration.Observe(duration.Seconds())
	if err != nil {
		milvusSearchErrorsTotal.Inc()
	}
}

func ObserveAICall(kind string, duration time.Duration, err error) {
	status := "ok"
	if err != nil {
		status = "failed"
		aiCallErrorsTotal.WithLabelValues(kind).Inc()
	}
	aiCallsTotal.WithLabelValues(kind, status).Inc()
	aiCallDuration.WithLabelValues(kind).Observe(duration.Seconds())
}

func IncWorkerTask(taskType, status string) {
	workerTasksTotal.WithLabelValues(taskType, status).Inc()
}
