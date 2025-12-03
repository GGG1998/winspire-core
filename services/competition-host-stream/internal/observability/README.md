# Observability Package

CloudWatch custom metrics integration for monitoring Winspire microservices.

## Metrics Published

### SSE Metrics
- **SSEActiveConnections**: Current number of SSE connections on this instance
- **SSEMessagesSent**: Total messages sent to clients
- **SSETotalConnections**: Lifetime total connections
- **SSEAvgConnectionDuration**: Average duration clients stay connected

### HTTP Metrics
- **HTTPRequestCount**: Total HTTP requests processed
- **HTTPErrorCount**: Total HTTP errors (4xx, 5xx)
- **HTTPAvgLatency**: Average request latency in milliseconds
- **HTTPP95Latency**: 95th percentile request latency

### Database Metrics
- **DBConnectionPoolActive**: Active database connections
- **DBConnectionPoolIdle**: Idle database connections
- **DBAvgQueryTime**: Average query execution time

## Usage

### Initialize in main.go

```go
package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/winspire/competition-host-stream/internal/observability"
	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
)

func main() {
	ctx := context.Background()
	logger := slog.Default()

	// Create SSE metrics collector
	sseMetrics := ssebroker.NewMetricsCollector()

	// Create CloudWatch publisher
	cwPublisher, err := observability.NewCloudWatchPublisher(
		ctx,
		"Winspire",                    // namespace
		"competition-host-stream",     // service name
		logger,
	)
	if err != nil {
		logger.Error("failed to create CloudWatch publisher", "error", err)
		// Continue without CloudWatch (will log errors)
	}

	// Create metrics collector
	metricsCollector := observability.NewMetricsCollector(
		sseMetrics,
		dbPool,
		cwPublisher,
		logger,
		60*time.Second, // publish every 60 seconds
	)

	// Start collecting and publishing metrics
	metricsCollector.Start(ctx)
	defer metricsCollector.Stop()

	// ... rest of application
}
```

### Integrate with HTTP Middleware

```go
// In your HTTP handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// ... process request ...
	
	latency := time.Since(start)
	isError := statusCode >= 400
	
	h.metricsCollector.HTTPMetrics().RecordRequest(latency, isError)
}
```

### Integrate with SSE Broker

```go
// SSE broker already has built-in metrics
broker := ssebroker.NewRedisBroker(100, redisClient, logger)

// Metrics are automatically collected
// - ConnectionOpened() called when client connects
// - ConnectionClosed() called when client disconnects
// - MessageSent() called when message is published
```

## CloudWatch Dashboard

Example CloudWatch dashboard configuration:

```json
{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["Winspire", "SSEActiveConnections", {"stat": "Average"}],
          [".", "SSETotalConnections", {"stat": "Sum"}]
        ],
        "period": 60,
        "stat": "Average",
        "region": "us-east-1",
        "title": "SSE Connections"
      }
    },
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["Winspire", "HTTPAvgLatency", {"stat": "Average"}],
          [".", "HTTPP95Latency", {"stat": "Average"}]
        ],
        "period": 60,
        "stat": "Average",
        "region": "us-east-1",
        "title": "HTTP Latency"
      }
    }
  ]
}
```

## Auto-Scaling Based on Custom Metrics

### ECS Auto-Scaling Policy

```hcl
resource "aws_appautoscaling_policy" "sse_connections" {
  name               = "sse-connections-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    customized_metric_specification {
      metric_name = "SSEActiveConnections"
      namespace   = "Winspire"
      statistic   = "Average"
      
      dimensions {
        name  = "Service"
        value = "competition-host-stream"
      }
    }
    
    target_value       = 500  # Scale when > 500 connections per instance
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}
```

## CloudWatch Alarms

### High SSE Connection Count

```hcl
resource "aws_cloudwatch_metric_alarm" "high_sse_connections" {
  alarm_name          = "high-sse-connections"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "SSEActiveConnections"
  namespace           = "Winspire"
  period              = "60"
  statistic           = "Average"
  threshold           = "800"
  alarm_description   = "SSE connections approaching capacity"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    Service = "competition-host-stream"
  }
}
```

### High Error Rate

```hcl
resource "aws_cloudwatch_metric_alarm" "high_error_rate" {
  alarm_name          = "high-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  
  metric_query {
    id          = "error_rate"
    expression  = "errors / requests * 100"
    label       = "Error Rate"
    return_data = true
  }
  
  metric_query {
    id = "errors"
    metric {
      metric_name = "HTTPErrorCount"
      namespace   = "Winspire"
      period      = "60"
      stat        = "Sum"
      dimensions = {
        Service = "competition-host-stream"
      }
    }
  }
  
  metric_query {
    id = "requests"
    metric {
      metric_name = "HTTPRequestCount"
      namespace   = "Winspire"
      period      = "60"
      stat        = "Sum"
      dimensions = {
        Service = "competition-host-stream"
      }
    }
  }
  
  threshold         = "5"  # 5% error rate
  alarm_description = "HTTP error rate exceeds 5%"
  alarm_actions     = [aws_sns_topic.alerts.arn]
}
```

## Cost Considerations

CloudWatch custom metrics cost:
- First 10,000 metrics: $0.30 per metric per month
- Next 240,000 metrics: $0.10 per metric per month
- API requests: $0.01 per 1,000 PutMetricData requests

For this implementation (6 metrics × 1 service):
- **Metrics**: 6 × $0.30 = $1.80/month
- **API requests** (60s interval): ~43,200/month = $0.43/month
- **Total**: ~$2.23/month

## Troubleshooting

### Metrics not appearing in CloudWatch

1. Check IAM permissions for ECS task role:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudwatch:PutMetricData"
      ],
      "Resource": "*"
    }
  ]
}
```

2. Check CloudWatch publisher logs:
```bash
aws logs tail /ecs/dev/competition-host-stream --filter-pattern="cloudwatch_publisher" --follow
```

3. Verify AWS SDK configuration:
```bash
# In ECS task
aws cloudwatch list-metrics --namespace Winspire
```


