# ALB Module

Application Load Balancer module for Winspire microservices with path-based routing.

## Features

- **Path-based routing**: Route traffic to different services based on URL path
- **SSL/TLS termination**: HTTPS support with ACM certificates
- **Health checks**: Built-in health check configuration
- **Security groups**: Pre-configured security groups for ALB
- **Sticky sessions**: Support for SSE connections

## Usage

```hcl
module "alb" {
  source = "../../modules/alb"

  environment        = "dev"
  vpc_id             = module.vpc.vpc_id
  public_subnet_ids  = module.vpc.public_subnet_ids
  certificate_arn    = aws_acm_certificate.main.arn

  tags = {
    Project = "Winspire"
    Team    = "Platform"
  }
}
```

## Routing Configuration

Add listener rules in your environment configuration:

```hcl
resource "aws_lb_listener_rule" "competition_host_stream" {
  listener_arn = module.alb.http_listener_arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = module.competition_host_stream.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/stream/*", "/v1/cups/*", "/v1/tournaments/*"]
    }
  }
}
```

## Sticky Sessions for SSE

```hcl
resource "aws_lb_target_group" "competition_host_stream" {
  name     = "${var.environment}-competition-host-stream"
  port     = 8086
  protocol = "HTTP"
  vpc_id   = var.vpc_id
  target_type = "ip"

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 3600  # 1 hour
    enabled         = true
  }
}
```

## Outputs

- `alb_dns_name`: DNS name for Route53 configuration
- `alb_arn`: ARN for listener rules
- `alb_security_group_id`: Security group for ECS services

