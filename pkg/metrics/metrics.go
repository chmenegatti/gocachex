// Package metrics provides Prometheus metrics collection for GoCacheX.
package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chmenegatti/gocachex/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector collects and exports metrics for GoCacheX.
type Collector struct {
	config config.PrometheusConfig

	// Operation metrics
	operationsTotal   *prometheus.CounterVec
	operationDuration *prometheus.HistogramVec

	// Cache hit/miss metrics
	cacheHitsTotal   *prometheus.CounterVec
	cacheMissesTotal *prometheus.CounterVec

	// Cache size metrics
	cacheSizeBytes *prometheus.GaugeVec
	cacheKeyCount  *prometheus.GaugeVec

	// Connection metrics
	activeConnections *prometheus.GaugeVec

	// Error metrics
	errorsTotal *prometheus.CounterVec

	// Registry
	registry *prometheus.Registry
}

// New creates a new metrics collector.
func New(cfg config.PrometheusConfig) *Collector {
	namespace := cfg.Namespace
	if namespace == "" {
		namespace = "gocachex"
	}

	subsystem := cfg.Subsystem
	if subsystem == "" {
		subsystem = "cache"
	}

	collector := &Collector{
		config:   cfg,
		registry: prometheus.NewRegistry(),
	}

	// Operation metrics
	collector.operationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "operations_total",
			Help:      "Total number of cache operations",
		},
		[]string{"operation", "backend", "status"},
	)

	collector.operationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "operation_duration_seconds",
			Help:      "Duration of cache operations in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation", "backend"},
	)

	// Hit/miss metrics
	collector.cacheHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "hits_total",
			Help:      "Total number of cache hits",
		},
		[]string{"backend", "level"},
	)

	collector.cacheMissesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "misses_total",
			Help:      "Total number of cache misses",
		},
		[]string{"backend", "level"},
	)

	// Size metrics
	collector.cacheSizeBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "size_bytes",
			Help:      "Current cache size in bytes",
		},
		[]string{"backend", "level"},
	)

	collector.cacheKeyCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "key_count",
			Help:      "Current number of keys in cache",
		},
		[]string{"backend", "level"},
	)

	// Connection metrics
	collector.activeConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "active_connections",
			Help:      "Number of active connections",
		},
		[]string{"backend"},
	)

	// Error metrics
	collector.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total number of errors",
		},
		[]string{"operation", "backend", "error_type"},
	)

	// Register metrics
	collector.registry.MustRegister(
		collector.operationsTotal,
		collector.operationDuration,
		collector.cacheHitsTotal,
		collector.cacheMissesTotal,
		collector.cacheSizeBytes,
		collector.cacheKeyCount,
		collector.activeConnections,
		collector.errorsTotal,
	)

	return collector
}

// RecordOperation records a cache operation with its duration.
func (c *Collector) RecordOperation(operation string, duration time.Duration) {
	if c == nil {
		return
	}

	backend := "unknown"
	status := "success"

	c.operationsTotal.WithLabelValues(operation, backend, status).Inc()
	c.operationDuration.WithLabelValues(operation, backend).Observe(duration.Seconds())
}

// RecordOperationWithBackend records a cache operation with backend and status.
func (c *Collector) RecordOperationWithBackend(operation, backend, status string, duration time.Duration) {
	if c == nil {
		return
	}

	c.operationsTotal.WithLabelValues(operation, backend, status).Inc()
	c.operationDuration.WithLabelValues(operation, backend).Observe(duration.Seconds())
}

// RecordHit records a cache hit.
func (c *Collector) RecordHit(backend, level string) {
	if c == nil {
		return
	}

	c.cacheHitsTotal.WithLabelValues(backend, level).Inc()
}

// RecordMiss records a cache miss.
func (c *Collector) RecordMiss(backend, level string) {
	if c == nil {
		return
	}

	c.cacheMissesTotal.WithLabelValues(backend, level).Inc()
}

// UpdateCacheSize updates the cache size metric.
func (c *Collector) UpdateCacheSize(backend, level string, sizeBytes int64) {
	if c == nil {
		return
	}

	c.cacheSizeBytes.WithLabelValues(backend, level).Set(float64(sizeBytes))
}

// UpdateKeyCount updates the key count metric.
func (c *Collector) UpdateKeyCount(backend, level string, count int64) {
	if c == nil {
		return
	}

	c.cacheKeyCount.WithLabelValues(backend, level).Set(float64(count))
}

// UpdateActiveConnections updates the active connections metric.
func (c *Collector) UpdateActiveConnections(backend string, count int64) {
	if c == nil {
		return
	}

	c.activeConnections.WithLabelValues(backend).Set(float64(count))
}

// RecordError records an error.
func (c *Collector) RecordError(operation, backend, errorType string) {
	if c == nil {
		return
	}

	c.errorsTotal.WithLabelValues(operation, backend, errorType).Inc()
}

// StartMetricsServer starts the Prometheus metrics HTTP server.
func (c *Collector) StartMetricsServer() error {
	if c == nil {
		return fmt.Errorf("metrics collector is nil")
	}

	if !c.config.Enabled {
		return nil
	}

	port := c.config.Port
	if port == 0 {
		port = 8080
	}

	path := c.config.Path
	if path == "" {
		path = "/metrics"
	}

	handler := promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	http.Handle(path, handler)

	addr := ":" + strconv.Itoa(port)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			// Log error in production, for now just ignore
			_ = err
		}
	}()

	return nil
}

// GetRegistry returns the Prometheus registry for external use.
func (c *Collector) GetRegistry() *prometheus.Registry {
	if c == nil {
		return nil
	}
	return c.registry
}

// Close cleans up the metrics collector.
func (c *Collector) Close() error {
	// Nothing to clean up for now
	return nil
}
