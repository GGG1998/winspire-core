package ssebroker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector collects SSE metrics for CloudWatch
type MetricsCollector struct {
	mu                  sync.RWMutex
	activeConnections   int64
	totalConnections    int64
	totalDisconnections int64
	messagesSent        int64

	// Connection duration histogram (in seconds)
	connectionDurations []time.Duration
	maxHistogramSize    int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		connectionDurations: make([]time.Duration, 0, 1000),
		maxHistogramSize:    1000,
	}
}

// ConnectionOpened records a new SSE connection
func (m *MetricsCollector) ConnectionOpened() {
	atomic.AddInt64(&m.activeConnections, 1)
	atomic.AddInt64(&m.totalConnections, 1)
}

// ConnectionClosed records a closed SSE connection
func (m *MetricsCollector) ConnectionClosed(duration time.Duration) {
	atomic.AddInt64(&m.activeConnections, -1)
	atomic.AddInt64(&m.totalDisconnections, 1)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to histogram (bounded size)
	if len(m.connectionDurations) < m.maxHistogramSize {
		m.connectionDurations = append(m.connectionDurations, duration)
	}
}

// MessageSent records a message sent to a client
func (m *MetricsCollector) MessageSent() {
	atomic.AddInt64(&m.messagesSent, 1)
}

// GetMetrics returns current metrics snapshot
func (m *MetricsCollector) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return Metrics{
		ActiveConnections:     atomic.LoadInt64(&m.activeConnections),
		TotalConnections:      atomic.LoadInt64(&m.totalConnections),
		TotalDisconnections:   atomic.LoadInt64(&m.totalDisconnections),
		MessagesSent:          atomic.LoadInt64(&m.messagesSent),
		AvgConnectionDuration: m.calculateAvgDuration(),
	}
}

// calculateAvgDuration calculates average connection duration
func (m *MetricsCollector) calculateAvgDuration() time.Duration {
	if len(m.connectionDurations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range m.connectionDurations {
		total += d
	}

	return total / time.Duration(len(m.connectionDurations))
}

// Reset resets all metrics (useful for testing)
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.activeConnections, 0)
	atomic.StoreInt64(&m.totalConnections, 0)
	atomic.StoreInt64(&m.totalDisconnections, 0)
	atomic.StoreInt64(&m.messagesSent, 0)
	m.connectionDurations = make([]time.Duration, 0, m.maxHistogramSize)
}

// Metrics represents a snapshot of SSE metrics
type Metrics struct {
	ActiveConnections     int64         `json:"active_connections"`
	TotalConnections      int64         `json:"total_connections"`
	TotalDisconnections   int64         `json:"total_disconnections"`
	MessagesSent          int64         `json:"messages_sent"`
	AvgConnectionDuration time.Duration `json:"avg_connection_duration"`
}

// PublishMetricsToCloudWatch publishes metrics to CloudWatch (placeholder)
// This should be called periodically (e.g., every 60 seconds)
func (m *MetricsCollector) PublishMetricsToCloudWatch(ctx context.Context, namespace, serviceName string) error {
	metrics := m.GetMetrics()

	// TODO: Implement actual CloudWatch publishing
	// For now, this is a placeholder that would use AWS SDK:
	//
	// cloudwatchClient.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
	//     Namespace: aws.String(namespace),
	//     MetricData: []types.MetricDatum{
	//         {
	//             MetricName: aws.String("SSEActiveConnections"),
	//             Value:      aws.Float64(float64(metrics.ActiveConnections)),
	//             Unit:       types.StandardUnitCount,
	//             Dimensions: []types.Dimension{
	//                 {Name: aws.String("Service"), Value: aws.String(serviceName)},
	//             },
	//         },
	//         // ... more metrics
	//     },
	// })

	_ = metrics // Placeholder to avoid unused variable error
	return nil
}
