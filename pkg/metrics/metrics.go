package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "openwan_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "openwan_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// Database metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "openwan_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Cache metrics
	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_name"},
	)

	cacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_name"},
	)

	// File operations metrics
	fileUploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_file_uploads_total",
			Help: "Total number of file uploads",
		},
		[]string{"file_type", "status"},
	)

	fileUploadBytes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_file_upload_bytes_total",
			Help: "Total bytes uploaded",
		},
		[]string{"file_type"},
	)

	fileDownloadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_file_downloads_total",
			Help: "Total number of file downloads",
		},
		[]string{"file_type"},
	)

	// Transcoding metrics
	transcodingJobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_transcoding_jobs_total",
			Help: "Total number of transcoding jobs",
		},
		[]string{"status"},
	)

	transcodingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "openwan_transcoding_duration_seconds",
			Help:    "Transcoding job duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600},
		},
	)

	transcodingJobsInProgress = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "openwan_transcoding_jobs_in_progress",
			Help: "Number of transcoding jobs currently in progress",
		},
	)

	// Queue metrics
	queueMessagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_queue_messages_sent_total",
			Help: "Total number of messages sent to queue",
		},
		[]string{"queue_name"},
	)

	queueMessagesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_queue_messages_received_total",
			Help: "Total number of messages received from queue",
		},
		[]string{"queue_name"},
	)

	queueMessagesFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_queue_messages_failed_total",
			Help: "Total number of failed message processing",
		},
		[]string{"queue_name"},
	)

	// Search metrics
	searchQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_search_queries_total",
			Help: "Total number of search queries",
		},
		[]string{"search_type"},
	)

	searchQueryDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "openwan_search_query_duration_seconds",
			Help:    "Search query duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1, 2, 5},
		},
	)

	// Authentication metrics
	authAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openwan_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"result"},
	)
)

// PrometheusMiddleware instruments HTTP requests with Prometheus metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		httpRequestsInFlight.Inc()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
		httpRequestsInFlight.Dec()
	}
}

// RecordDBQuery records database query metrics
func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCacheHit records cache hit
func RecordCacheHit(cacheName string) {
	cacheHitsTotal.WithLabelValues(cacheName).Inc()
}

// RecordCacheMiss records cache miss
func RecordCacheMiss(cacheName string) {
	cacheMissesTotal.WithLabelValues(cacheName).Inc()
}

// RecordFileUpload records file upload metrics
func RecordFileUpload(fileType, status string, bytes int64) {
	fileUploadsTotal.WithLabelValues(fileType, status).Inc()
	if status == "success" {
		fileUploadBytes.WithLabelValues(fileType).Add(float64(bytes))
	}
}

// RecordFileDownload records file download
func RecordFileDownload(fileType string) {
	fileDownloadsTotal.WithLabelValues(fileType).Inc()
}

// RecordTranscodingJob records transcoding job metrics
func RecordTranscodingJob(status string) {
	transcodingJobsTotal.WithLabelValues(status).Inc()
}

// RecordTranscodingDuration records transcoding duration
func RecordTranscodingDuration(duration time.Duration) {
	transcodingDuration.Observe(duration.Seconds())
}

// IncTranscodingJobsInProgress increments transcoding jobs in progress
func IncTranscodingJobsInProgress() {
	transcodingJobsInProgress.Inc()
}

// DecTranscodingJobsInProgress decrements transcoding jobs in progress
func DecTranscodingJobsInProgress() {
	transcodingJobsInProgress.Dec()
}

// RecordQueueMessageSent records message sent to queue
func RecordQueueMessageSent(queueName string) {
	queueMessagesSent.WithLabelValues(queueName).Inc()
}

// RecordQueueMessageReceived records message received from queue
func RecordQueueMessageReceived(queueName string) {
	queueMessagesReceived.WithLabelValues(queueName).Inc()
}

// RecordQueueMessageFailed records failed message processing
func RecordQueueMessageFailed(queueName string) {
	queueMessagesFailed.WithLabelValues(queueName).Inc()
}

// RecordSearchQuery records search query metrics
func RecordSearchQuery(searchType string, duration time.Duration) {
	searchQueriesTotal.WithLabelValues(searchType).Inc()
	searchQueryDuration.Observe(duration.Seconds())
}

// RecordAuthAttempt records authentication attempt
func RecordAuthAttempt(result string) {
	authAttemptsTotal.WithLabelValues(result).Inc()
}
