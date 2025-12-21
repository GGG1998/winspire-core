#!/bin/bash
# LocalStack ALB Initialization Script
# This script runs automatically when LocalStack starts (via /etc/localstack/init/ready.d/)
# Creates VPC, subnets, security group, ALB, target groups, and routing rules
#
# Routing Configuration (matches Traefik config):
#   - /v1/games*, /v1/g*, /v1/admin* -> game-management:8085 (priority 100)
#   - /v1/matchmaking* -> matchmaking:8081 (priority 101)
#   - /v1/* -> competition:8089 (default action)

set -e

echo "============================================"
echo "LocalStack ALB Setup Script"
echo "============================================"

# Wait for LocalStack to be fully ready
echo "Waiting for LocalStack to be ready..."
until awslocal sts get-caller-identity &>/dev/null; do
    sleep 1
done
echo "LocalStack is ready!"

# Configuration
REGION="eu-central-1"
VPC_CIDR="10.0.0.0/16"
SUBNET_1_CIDR="10.0.1.0/24"
SUBNET_2_CIDR="10.0.2.0/24"

# ==============================================================================
# Create VPC and Networking
# ==============================================================================
echo ""
echo "[1/7] Creating VPC..."
VPC_ID=$(awslocal ec2 create-vpc \
    --cidr-block "$VPC_CIDR" \
    --query 'Vpc.VpcId' \
    --output text)
echo "Created VPC: $VPC_ID"

# Enable DNS hostnames
awslocal ec2 modify-vpc-attribute --vpc-id "$VPC_ID" --enable-dns-hostnames

echo ""
echo "[2/7] Creating subnets..."
SUBNET_1=$(awslocal ec2 create-subnet \
    --vpc-id "$VPC_ID" \
    --cidr-block "$SUBNET_1_CIDR" \
    --availability-zone "${REGION}a" \
    --query 'Subnet.SubnetId' \
    --output text)
echo "Created subnet 1: $SUBNET_1 (${REGION}a)"

SUBNET_2=$(awslocal ec2 create-subnet \
    --vpc-id "$VPC_ID" \
    --cidr-block "$SUBNET_2_CIDR" \
    --availability-zone "${REGION}b" \
    --query 'Subnet.SubnetId' \
    --output text)
echo "Created subnet 2: $SUBNET_2 (${REGION}b)"

# ==============================================================================
# Create Security Group
# ==============================================================================
echo ""
echo "[3/7] Creating security group..."
SG_ID=$(awslocal ec2 create-security-group \
    --group-name winspire-alb-sg \
    --description "Winspire ALB Security Group" \
    --vpc-id "$VPC_ID" \
    --query 'GroupId' \
    --output text)

awslocal ec2 authorize-security-group-ingress \
    --group-id "$SG_ID" \
    --protocol tcp \
    --port 80 \
    --cidr 0.0.0.0/0

awslocal ec2 authorize-security-group-ingress \
    --group-id "$SG_ID" \
    --protocol tcp \
    --port 443 \
    --cidr 0.0.0.0/0

echo "Created security group: $SG_ID"

# ==============================================================================
# Create Application Load Balancer
# ==============================================================================
echo ""
echo "[4/7] Creating Application Load Balancer..."
ALB_ARN=$(awslocal elbv2 create-load-balancer \
    --name winspire-local-alb \
    --subnets "$SUBNET_1" "$SUBNET_2" \
    --security-groups "$SG_ID" \
    --scheme internet-facing \
    --type application \
    --query 'LoadBalancers[0].LoadBalancerArn' \
    --output text)
echo "Created ALB: $ALB_ARN"

ALB_DNS=$(awslocal elbv2 describe-load-balancers \
    --names winspire-local-alb \
    --query 'LoadBalancers[0].DNSName' \
    --output text)
echo "ALB DNS: $ALB_DNS"

# ==============================================================================
# Create Target Groups
# ==============================================================================
echo ""
echo "[5/7] Creating target groups..."

# Competition service (default target - catches /v1/*)
TG_COMPETITION=$(awslocal elbv2 create-target-group \
    --name competition-tg \
    --protocol HTTP \
    --port 8089 \
    --vpc-id "$VPC_ID" \
    --target-type ip \
    --health-check-path /healthz \
    --health-check-interval-seconds 30 \
    --health-check-timeout-seconds 5 \
    --healthy-threshold-count 2 \
    --unhealthy-threshold-count 2 \
    --query 'TargetGroups[0].TargetGroupArn' \
    --output text)
echo "Created Competition TG: $TG_COMPETITION"

# Game Management service (/v1/games, /v1/g, /v1/admin)
TG_GAME_MGMT=$(awslocal elbv2 create-target-group \
    --name game-management-tg \
    --protocol HTTP \
    --port 8085 \
    --vpc-id "$VPC_ID" \
    --target-type ip \
    --health-check-path /healthz \
    --health-check-interval-seconds 30 \
    --health-check-timeout-seconds 5 \
    --healthy-threshold-count 2 \
    --unhealthy-threshold-count 2 \
    --query 'TargetGroups[0].TargetGroupArn' \
    --output text)
echo "Created Game Management TG: $TG_GAME_MGMT"

# Matchmaking service (/v1/matchmaking)
TG_MATCHMAKING=$(awslocal elbv2 create-target-group \
    --name matchmaking-tg \
    --protocol HTTP \
    --port 8081 \
    --vpc-id "$VPC_ID" \
    --target-type ip \
    --health-check-path /health \
    --health-check-interval-seconds 30 \
    --health-check-timeout-seconds 5 \
    --healthy-threshold-count 2 \
    --unhealthy-threshold-count 2 \
    --query 'TargetGroups[0].TargetGroupArn' \
    --output text)
echo "Created Matchmaking TG: $TG_MATCHMAKING"

# ==============================================================================
# Create Listener and Routing Rules
# ==============================================================================
echo ""
echo "[6/7] Creating listener and routing rules..."

# Create default listener with competition as default action
LISTENER_ARN=$(awslocal elbv2 create-listener \
    --load-balancer-arn "$ALB_ARN" \
    --protocol HTTP \
    --port 80 \
    --default-actions Type=forward,TargetGroupArn="$TG_COMPETITION" \
    --query 'Listeners[0].ListenerArn' \
    --output text)
echo "Created listener: $LISTENER_ARN"

# Rule 1: /v1/games*, /v1/g*, /v1/admin* -> game-management (priority 100)
awslocal elbv2 create-rule \
    --listener-arn "$LISTENER_ARN" \
    --priority 100 \
    --conditions '[{"Field":"path-pattern","PathPatternConfig":{"Values":["/v1/games*","/v1/g/*","/v1/admin*"]}}]' \
    --actions Type=forward,TargetGroupArn="$TG_GAME_MGMT" \
    > /dev/null
echo "Created rule: /v1/games*, /v1/g/*, /v1/admin* -> game-management (priority 100)"

# Rule 2: /v1/matchmaking* -> matchmaking (priority 101)
awslocal elbv2 create-rule \
    --listener-arn "$LISTENER_ARN" \
    --priority 101 \
    --conditions '[{"Field":"path-pattern","PathPatternConfig":{"Values":["/v1/matchmaking*"]}}]' \
    --actions Type=forward,TargetGroupArn="$TG_MATCHMAKING" \
    > /dev/null
echo "Created rule: /v1/matchmaking* -> matchmaking (priority 101)"

echo "Default rule: /v1/* -> competition"

# ==============================================================================
# Export Configuration
# ==============================================================================
echo ""
echo "[7/7] Saving configuration..."

# Save config for other scripts
cat > /tmp/alb-config.env << EOF
VPC_ID=$VPC_ID
SUBNET_1=$SUBNET_1
SUBNET_2=$SUBNET_2
SG_ID=$SG_ID
ALB_ARN=$ALB_ARN
ALB_DNS=$ALB_DNS
LISTENER_ARN=$LISTENER_ARN
TG_COMPETITION=$TG_COMPETITION
TG_GAME_MGMT=$TG_GAME_MGMT
TG_MATCHMAKING=$TG_MATCHMAKING
EOF

echo "Configuration saved to /tmp/alb-config.env"

# ==============================================================================
# Summary
# ==============================================================================
echo ""
echo "============================================"
echo "ALB Setup Complete!"
echo "============================================"
echo ""
echo "Routing Configuration:"
echo "  /v1/games*, /v1/g/*, /v1/admin* -> game-management:8085"
echo "  /v1/matchmaking*                -> matchmaking:8081"
echo "  /v1/* (default)                 -> competition:8089"
echo ""
echo "ALB DNS: $ALB_DNS"
echo ""
echo "Note: Target registration will be handled by the alb-registrar container"
echo "============================================"
