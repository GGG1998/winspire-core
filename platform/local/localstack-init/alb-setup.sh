#!/bin/bash
# LocalStack ALB Initialization Script
# This script runs automatically when LocalStack starts (via /etc/localstack/init/ready.d/)
# Creates VPC, subnets, security group, ALB, target groups, and routing rules
#
# MONOLITH MODE: Tournament service merged into matchmaking
#
# Routing Configuration:
#   - /v1/* -> matchmaking:8081 (all APIs - game management + tournament + matchmaking)

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

# Create S3 buckets
echo "Creating S3 bucket: games"
awslocal s3 mb s3://games || echo "Bucket games already exists"

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

# MONOLITH MODE: Only matchmaking service (includes game-management + tournament)
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
echo "Created Matchmaking TG: $TG_MATCHMAKING (monolith - all APIs)"

# ==============================================================================
# Create Listener and Routing Rules
# ==============================================================================
echo ""
echo "[6/7] Creating listener and routing rules..."

# MONOLITH MODE: Create default listener with matchmaking as default action (all APIs)
LISTENER_ARN=$(awslocal elbv2 create-listener \
    --load-balancer-arn "$ALB_ARN" \
    --protocol HTTP \
    --port 80 \
    --default-actions Type=forward,TargetGroupArn="$TG_MATCHMAKING" \
    --query 'Listeners[0].ListenerArn' \
    --output text)
echo "Created listener: $LISTENER_ARN"

echo "Default rule: /v1/* -> matchmaking (monolith - all APIs)"

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
TG_MATCHMAKING=$TG_MATCHMAKING
EOF

echo "Configuration saved to /tmp/alb-config.env"

# ==============================================================================
# Summary
# ==============================================================================
echo ""
echo "============================================"
echo "ALB Setup Complete! (Monolith Mode)"
echo "============================================"
echo ""
echo "Routing Configuration:"
echo "  /v1/* (all APIs) -> matchmaking:8081"
echo ""
echo "Included APIs:"
echo "  - Game Management (/v1/games, /v1/g/*, /v1/admin)"
echo "  - Tournament (/v1/hosts, /v1/:hostId/tournaments)"
echo "  - Matchmaking (/v1/matchmaking/*)"
echo ""
echo "ALB DNS: $ALB_DNS"
echo ""
echo "Note: Target registration will be handled by the alb-registrar container"
echo "============================================"
