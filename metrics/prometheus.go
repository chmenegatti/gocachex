package metrics

import (
	"context"
	"time"

	"github.com/chmenegatti/gocachex"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_name"},
	)
	cacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_name"},
	)
	cacheSets = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_set_total",
			Help: "Total number of cache sets",
		},
		[]string{"cache_name"},
	)
	cacheLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_latency_seconds",
			Help:    "Cache operation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"cache_name", "operation"},
	)
)

// PrometheusCache is a wrapper that adds Prometheus metrics to any Cache implementation.
type PrometheusCache[T any] struct {
	cache gocachex.Cache[T]
	name  string
}

// NewPrometheusCache wraps an existing cache with Prometheus metrics tracking.
func NewPrometheusCache[T any](cache gocachex.Cache[T], cacheName string) *PrometheusCache[T] {
	return &PrometheusCache[T]{
		cache: cache,
		name:  cacheName,
	}
}

// Get retrieves a value and records the metric.
func (mw *PrometheusCache[T]) Get(ctx context.Context, key string) (T, error) {
	start := time.Now()
	val, err := mw.cache.Get(ctx, key)
	duration := time.Since(start).Seconds()

	cacheLatency.WithLabelValues(mw.name, "get").Observe(duration)

	if err == nil {
		cacheHits.WithLabelValues(mw.name).Inc()
	} else if err == gocachex.ErrCacheMiss {
		cacheMisses.WithLabelValues(mw.name).Inc()
	}

	return val, err
}

// Set stores a value and records the metric.
func (mw *PrometheusCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	start := time.Now()
	err := mw.cache.Set(ctx, key, value, ttl)
	duration := time.Since(start).Seconds()

	cacheLatency.WithLabelValues(mw.name, "set").Observe(duration)
	if err == nil {
		cacheSets.WithLabelValues(mw.name).Inc()
	}

	return err
}

// Delete removes a value and records latency.
func (mw *PrometheusCache[T]) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := mw.cache.Delete(ctx, key)
	duration := time.Since(start).Seconds()

	cacheLatency.WithLabelValues(mw.name, "delete").Observe(duration)
	return err
}

// Exists checks if a key exists and records latency.
func (mw *PrometheusCache[T]) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	val, err := mw.cache.Exists(ctx, key)
	duration := time.Since(start).Seconds()

	cacheLatency.WithLabelValues(mw.name, "exists").Observe(duration)
	return val, err
}

// Clear purges all keys and records latency.
func (mw *PrometheusCache[T]) Clear(ctx context.Context) error {
	start := time.Now()
	err := mw.cache.Clear(ctx)
	duration := time.Since(start).Seconds()

	cacheLatency.WithLabelValues(mw.name, "clear").Observe(duration)
	return err
}

// Close closes the underlying cache.
func (mw *PrometheusCache[T]) Close() error {
	return mw.cache.Close()
}

// compiler check
var _ gocachex.Cache[string] = (*PrometheusCache[string])(nil)
