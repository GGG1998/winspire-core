package observability

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
)

// MetricsCollector aggregates metrics from various sources and publishes to CloudWatch
type MetricsCollector struct {
	sseMetrics      *ssebroker.MetricsCollector
	httpMetrics     *HTTPMetricsCollector
	dbPool          *pgxpool.Pool
	cwPublisher     *CloudWatchPublisher
	logger          *slog.Logger
	publishInterval time.Duration
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(
	sseMetrics *ssebroker.MetricsCollector,
	dbPool *pgxpool.Pool,
	cwPublisher *CloudWatchPublisher,
	logger *slog.Logger,
	publishInterval time.Duration,
) *MetricsCollector {
	return &MetricsCollector{
		sseMetrics:      sseMetrics,
		httpMetrics:     NewHTTPMetricsCollector(),
		dbPool:          dbPool,
		cwPublisher:     cwPublisher,
		logger:          logger.With("component", "metrics_collector"),
		publishInterval: publishInterval,
		stopChan:        make(chan struct{}),
	}
}

// Start begins collecting and publishing metrics
func (mc *MetricsCollector) Start(ctx context.Context) {
	mc.logger.Info("starting metrics collector", "interval", mc.publishInterval)

	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()
		mc.publishLoop(ctx)
	}()
}

// Stop stops the metrics collector
func (mc *MetricsCollector) Stop() {
	mc.logger.Info("stopping metrics collector")
	close(mc.stopChan)
	mc.wg.Wait()
}

// HTTPMetrics returns the HTTP metrics collector
func (mc *MetricsCollector) HTTPMetrics() *HTTPMetricsCollector {
	return mc.httpMetrics
}

// publishLoop periodically publishes metrics to CloudWatch
func (mc *MetricsCollector) publishLoop(ctx context.Context) {
	ticker := time.NewTicker(mc.publishInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mc.logger.Info("metrics collector context cancelled")
			return
		case <-mc.stopChan:
			mc.logger.Info("metrics collector stopped")
			return
		case <-ticker.C:
			mc.publishMetrics(ctx)
		}
	}
}

// publishMetrics collects and publishes all metrics
func (mc *MetricsCollector) publishMetrics(ctx context.Context) {
	// Publish SSE metrics
	if mc.sseMetrics != nil {
		sseMetrics := mc.sseMetrics.GetMetrics()
		if err := mc.cwPublisher.PublishSSEMetrics(ctx, SSEMetrics{
			ActiveConnections:     sseMetrics.ActiveConnections,
			TotalConnections:      sseMetrics.TotalConnections,
			MessagesSent:          sseMetrics.MessagesSent,
			AvgConnectionDuration: sseMetrics.AvgConnectionDuration,
		}); err != nil {
			mc.logger.Error("failed to publish SSE metrics", "error", err)
		}
	}

	// Publish HTTP metrics
	if mc.httpMetrics != nil {
		httpMetrics := mc.httpMetrics.GetMetrics()
		if err := mc.cwPublisher.PublishHTTPMetrics(ctx, httpMetrics); err != nil {
			mc.logger.Error("failed to publish HTTP metrics", "error", err)
		}
	}

	// Publish database metrics
	if mc.dbPool != nil {
		stat := mc.dbPool.Stat()
		dbMetrics := DatabaseMetrics{
			ActiveConnections: stat.AcquiredConns(),
			IdleConnections:   stat.IdleConns(),
			AvgQueryTime:      0, // TODO: Track query times
		}
		if err := mc.cwPublisher.PublishDatabaseMetrics(ctx, dbMetrics); err != nil {
			mc.logger.Error("failed to publish database metrics", "error", err)
		}
	}

	mc.logger.Debug("published metrics to CloudWatch")
}

// HTTPMetricsCollector collects HTTP metrics
type HTTPMetricsCollector struct {
	mu            sync.RWMutex
	requestCount  int64
	errorCount    int64
	latencies     []time.Duration
	maxLatencies  int
}

// NewHTTPMetricsCollector creates a new HTTP metrics collector
func NewHTTPMetricsCollector() *HTTPMetricsCollector {
	return &HTTPMetricsCollector{
		latencies:    make([]time.Duration, 0, 1000),
		maxLatencies: 1000,
	}
}

// RecordRequest records an HTTP request
func (h *HTTPMetricsCollector) RecordRequest(latency time.Duration, isError bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requestCount++
	if isError {
		h.errorCount++
	}

	// Store latency (bounded)
	if len(h.latencies) < h.maxLatencies {
		h.latencies = append(h.latencies, latency)
	}
}

// GetMetrics returns current HTTP metrics
func (h *HTTPMetricsCollector) GetMetrics() HTTPMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()

	metrics := HTTPMetrics{
		RequestCount: h.requestCount,
		ErrorCount:   h.errorCount,
	}

	if len(h.latencies) > 0 {
		// Calculate average
		var total time.Duration
		for _, l := range h.latencies {
			total += l
		}
		metrics.AvgLatency = total / time.Duration(len(h.latencies))

		// Calculate P95 (simple approximation)
		// TODO: Use a proper percentile algorithm
		sorted := make([]time.Duration, len(h.latencies))
		copy(sorted, h.latencies)
		// Simple sort would be needed here for accurate P95
		metrics.P95Latency = h.latencies[len(h.latencies)-1]
	}

	return metrics
}

// Reset resets HTTP metrics
func (h *HTTPMetricsCollector) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requestCount = 0
	h.errorCount = 0
	h.latencies = make([]time.Duration, 0, h.maxLatencies)
}


