package observability

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// CloudWatchPublisher publishes custom metrics to AWS CloudWatch
type CloudWatchPublisher struct {
	client      *cloudwatch.Client
	namespace   string
	serviceName string
	logger      *slog.Logger
}

// NewCloudWatchPublisher creates a new CloudWatch metrics publisher
func NewCloudWatchPublisher(ctx context.Context, namespace, serviceName string, logger *slog.Logger) (*CloudWatchPublisher, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &CloudWatchPublisher{
		client:      cloudwatch.NewFromConfig(cfg),
		namespace:   namespace,
		serviceName: serviceName,
		logger:      logger.With("component", "cloudwatch_publisher"),
	}, nil
}

// PublishSSEMetrics publishes SSE-related metrics to CloudWatch
func (p *CloudWatchPublisher) PublishSSEMetrics(ctx context.Context, metrics SSEMetrics) error {
	now := time.Now()

	metricData := []types.MetricDatum{
		{
			MetricName: aws.String("SSEActiveConnections"),
			Value:      aws.Float64(float64(metrics.ActiveConnections)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
		{
			MetricName: aws.String("SSEMessagesSent"),
			Value:      aws.Float64(float64(metrics.MessagesSent)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
		{
			MetricName: aws.String("SSETotalConnections"),
			Value:      aws.Float64(float64(metrics.TotalConnections)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
		{
			MetricName: aws.String("SSEAvgConnectionDuration"),
			Value:      aws.Float64(float64(metrics.AvgConnectionDuration.Seconds())),
			Unit:       types.StandardUnitSeconds,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
	}

	input := &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(p.namespace),
		MetricData: metricData,
	}

	_, err := p.client.PutMetricData(ctx, input)
	if err != nil {
		p.logger.Error("failed to publish SSE metrics to CloudWatch",
			"error", err,
			"namespace", p.namespace,
		)
		return fmt.Errorf("failed to publish metrics: %w", err)
	}

	p.logger.Debug("published SSE metrics to CloudWatch",
		"active_connections", metrics.ActiveConnections,
		"messages_sent", metrics.MessagesSent,
	)

	return nil
}

// PublishHTTPMetrics publishes HTTP-related metrics to CloudWatch
func (p *CloudWatchPublisher) PublishHTTPMetrics(ctx context.Context, metrics HTTPMetrics) error {
	now := time.Now()

	metricData := []types.MetricDatum{
		{
			MetricName: aws.String("HTTPRequestCount"),
			Value:      aws.Float64(float64(metrics.RequestCount)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
		{
			MetricName: aws.String("HTTPErrorCount"),
			Value:      aws.Float64(float64(metrics.ErrorCount)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
		{
			MetricName: aws.String("HTTPAvgLatency"),
			Value:      aws.Float64(float64(metrics.AvgLatency.Milliseconds())),
			Unit:       types.StandardUnitMilliseconds,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
		{
			MetricName: aws.String("HTTPP95Latency"),
			Value:      aws.Float64(float64(metrics.P95Latency.Milliseconds())),
			Unit:       types.StandardUnitMilliseconds,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
	}

	input := &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(p.namespace),
		MetricData: metricData,
	}

	_, err := p.client.PutMetricData(ctx, input)
	if err != nil {
		p.logger.Error("failed to publish HTTP metrics to CloudWatch",
			"error", err,
			"namespace", p.namespace,
		)
		return fmt.Errorf("failed to publish metrics: %w", err)
	}

	p.logger.Debug("published HTTP metrics to CloudWatch",
		"request_count", metrics.RequestCount,
		"error_count", metrics.ErrorCount,
	)

	return nil
}

// PublishDatabaseMetrics publishes database-related metrics to CloudWatch
func (p *CloudWatchPublisher) PublishDatabaseMetrics(ctx context.Context, metrics DatabaseMetrics) error {
	now := time.Now()

	metricData := []types.MetricDatum{
		{
			MetricName: aws.String("DBConnectionPoolActive"),
			Value:      aws.Float64(float64(metrics.ActiveConnections)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
		{
			MetricName: aws.String("DBConnectionPoolIdle"),
			Value:      aws.Float64(float64(metrics.IdleConnections)),
			Unit:       types.StandardUnitCount,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
		{
			MetricName: aws.String("DBAvgQueryTime"),
			Value:      aws.Float64(float64(metrics.AvgQueryTime.Milliseconds())),
			Unit:       types.StandardUnitMilliseconds,
			Timestamp:  &now,
			Dimensions: []types.Dimension{
				{Name: aws.String("Service"), Value: aws.String(p.serviceName)},
			},
		},
	}

	input := &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(p.namespace),
		MetricData: metricData,
	}

	_, err := p.client.PutMetricData(ctx, input)
	if err != nil {
		p.logger.Error("failed to publish database metrics to CloudWatch",
			"error", err,
			"namespace", p.namespace,
		)
		return fmt.Errorf("failed to publish metrics: %w", err)
	}

	p.logger.Debug("published database metrics to CloudWatch",
		"active_connections", metrics.ActiveConnections,
		"idle_connections", metrics.IdleConnections,
	)

	return nil
}

// SSEMetrics represents SSE-related metrics
type SSEMetrics struct {
	ActiveConnections     int64
	TotalConnections      int64
	MessagesSent          int64
	AvgConnectionDuration time.Duration
}

// HTTPMetrics represents HTTP-related metrics
type HTTPMetrics struct {
	RequestCount int64
	ErrorCount   int64
	AvgLatency   time.Duration
	P95Latency   time.Duration
}

// DatabaseMetrics represents database-related metrics
type DatabaseMetrics struct {
	ActiveConnections int32
	IdleConnections   int32
	AvgQueryTime      time.Duration
}
