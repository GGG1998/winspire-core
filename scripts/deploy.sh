#!/bin/bash
# Winspire Deployment Script
# Builds Docker images for services and pushes them to AWS ECR
#
# Usage:
#   ./scripts/deploy.sh                           # Deploy all services to dev
#   ./scripts/deploy.sh -e prod -u                # Deploy to prod and update ECS
#   ./scripts/deploy.sh -s tournament,matchmaking # Deploy specific services

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Default values
AWS_REGION="${AWS_REGION:-eu-central-1}"
ENVIRONMENT="${ENVIRONMENT:-dev}"
UPDATE_ECS="${UPDATE_ECS:-false}"
SERVICES="${SERVICES:-tournament,matchmaking,game-management}"
DRY_RUN="${DRY_RUN:-false}"

# Service port mapping
declare -A SERVICE_PORTS=(
    ["tournament"]="8089"
    ["matchmaking"]="8081"
    ["game-management"]="8087"
)

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Builds and deploys Winspire microservices to AWS ECR.

Options:
    -e, --environment   Environment (dev|staging|prod) [default: dev]
    -r, --region        AWS region [default: eu-central-1]
    -s, --services      Comma-separated list of services to deploy
                        [default: tournament,matchmaking,game-management]
    -u, --update-ecs    Update ECS services after pushing images
    -d, --dry-run       Show what would be done without executing
    -h, --help          Show this help message

Examples:
    $0                                    # Deploy all services to dev
    $0 -e dev -s tournament,matchmaking  # Deploy specific services
    $0 -e prod -u                         # Deploy to prod and update ECS
    $0 --dry-run                          # Preview without executing

Environment Variables:
    AWS_REGION          Override default region
    AWS_PROFILE         Use specific AWS profile
    ENVIRONMENT         Override default environment
EOF
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -r|--region)
            AWS_REGION="$2"
            shift 2
            ;;
        -s|--services)
            SERVICES="$2"
            shift 2
            ;;
        -u|--update-ecs)
            UPDATE_ECS="true"
            shift
            ;;
        -d|--dry-run)
            DRY_RUN="true"
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate environment
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]]; then
    log_error "Invalid environment: $ENVIRONMENT. Must be dev, staging, or prod."
    exit 1
fi

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."

    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI not found. Install it: https://aws.amazon.com/cli/"
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Install Docker first."
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker daemon not running. Start Docker first."
        exit 1
    fi

    log_info "Prerequisites OK"
}

# Get AWS account info
get_aws_info() {
    log_step "Getting AWS account information..."

    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text 2>/dev/null)
    if [[ -z "$AWS_ACCOUNT_ID" ]]; then
        log_error "Failed to get AWS account ID. Check AWS credentials."
        log_error "Run: aws configure or set AWS_PROFILE"
        exit 1
    fi

    ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
    log_info "AWS Account: $AWS_ACCOUNT_ID"
    log_info "ECR Registry: $ECR_REGISTRY"
}

# Generate image tag
generate_tag() {
    GIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null | tr '/' '-' || echo "unknown")
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)

    IMAGE_TAG="${ENVIRONMENT}-${GIT_SHA}-${TIMESTAMP}"
    log_info "Image tag: $IMAGE_TAG"
}

# Authenticate with ECR
ecr_login() {
    log_step "Authenticating with ECR..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would authenticate with ECR"
        return
    fi

    aws ecr get-login-password --region "$AWS_REGION" | \
        docker login --username AWS --password-stdin "$ECR_REGISTRY"

    log_info "ECR authentication successful"
}

# Ensure ECR repository exists
ensure_ecr_repo() {
    local repo_name="$1"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would check/create ECR repo: $repo_name"
        return
    fi

    if ! aws ecr describe-repositories --repository-names "$repo_name" --region "$AWS_REGION" &>/dev/null; then
        log_warn "ECR repository does not exist: $repo_name"
        log_info "Creating ECR repository..."
        aws ecr create-repository \
            --repository-name "$repo_name" \
            --region "$AWS_REGION" \
            --image-scanning-configuration scanOnPush=true \
            --encryption-configuration encryptionType=AES256
        log_info "Created ECR repository: $repo_name"
    fi
}

# Build and push a service
build_and_push_service() {
    local service="$1"
    local service_dir="$PROJECT_ROOT/services/$service"
    local dockerfile="$service_dir/cmd/$service/Dockerfile"
    local ecr_repo="${ENVIRONMENT}-winspire-${service}"
    local full_image="${ECR_REGISTRY}/${ecr_repo}"
    local port="${SERVICE_PORTS[$service]:-8080}"

    log_step "Processing service: $service"

    # Verify service exists
    if [[ ! -d "$service_dir" ]]; then
        log_error "Service directory not found: $service_dir"
        return 1
    fi

    if [[ ! -f "$dockerfile" ]]; then
        log_error "Dockerfile not found: $dockerfile"
        return 1
    fi

    # Ensure ECR repository exists
    ensure_ecr_repo "$ecr_repo"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would build: $full_image:$IMAGE_TAG"
        log_info "[DRY-RUN] Would push: $full_image:$IMAGE_TAG"
        log_info "[DRY-RUN] Would push: $full_image:latest"
        return
    fi

    # Build Docker image
    log_info "Building Docker image for $service..."
    docker build \
        --platform linux/amd64 \
        -t "${full_image}:${IMAGE_TAG}" \
        -t "${full_image}:latest" \
        -t "${full_image}:${ENVIRONMENT}" \
        -f "$dockerfile" \
        "$service_dir"

    # Push to ECR
    log_info "Pushing image to ECR..."
    docker push "${full_image}:${IMAGE_TAG}"
    docker push "${full_image}:latest"
    docker push "${full_image}:${ENVIRONMENT}"

    log_info "Successfully pushed: ${full_image}:${IMAGE_TAG}"

    # Store for summary
    DEPLOYED_IMAGES+=("${full_image}:${IMAGE_TAG}")
}

# Update ECS service
update_ecs_service() {
    local service="$1"
    local ecs_cluster="${ENVIRONMENT}-winspire-cluster"
    local ecs_service="${ENVIRONMENT}-${service}"

    log_step "Updating ECS service: $ecs_service"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would update ECS service: $ecs_service in cluster $ecs_cluster"
        return
    fi

    # Check if service exists
    if ! aws ecs describe-services --cluster "$ecs_cluster" --services "$ecs_service" --region "$AWS_REGION" --query 'services[0].serviceName' --output text &>/dev/null; then
        log_warn "ECS service not found: $ecs_service (skipping)"
        return
    fi

    # Force new deployment
    aws ecs update-service \
        --cluster "$ecs_cluster" \
        --service "$ecs_service" \
        --force-new-deployment \
        --region "$AWS_REGION" \
        --output text > /dev/null

    log_info "ECS service update initiated for $service"
}

# Print summary
print_summary() {
    echo ""
    echo "============================================"
    echo "Deployment Summary"
    echo "============================================"
    echo "Environment: $ENVIRONMENT"
    echo "Region: $AWS_REGION"
    echo "Image Tag: $IMAGE_TAG"
    echo ""
    echo "Deployed Images:"
    for img in "${DEPLOYED_IMAGES[@]}"; do
        echo "  - $img"
    done

    if [[ "$UPDATE_ECS" == "true" ]]; then
        echo ""
        echo "ECS services have been updated."
        echo "Monitor deployments:"
        echo "  aws ecs describe-services --cluster ${ENVIRONMENT}-winspire-cluster --services ${SERVICE_ARRAY[*]/#/${ENVIRONMENT}-} --region $AWS_REGION"
    fi
    echo "============================================"
}

# Main execution
main() {
    echo ""
    echo "============================================"
    echo "Winspire Deployment Script"
    echo "============================================"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_warn "DRY-RUN MODE - No changes will be made"
    fi

    check_prerequisites
    get_aws_info
    generate_tag
    ecr_login

    # Convert services string to array
    IFS=',' read -ra SERVICE_ARRAY <<< "$SERVICES"
    DEPLOYED_IMAGES=()

    # Build and push each service
    for service in "${SERVICE_ARRAY[@]}"; do
        service=$(echo "$service" | xargs)  # Trim whitespace
        build_and_push_service "$service"
    done

    # Update ECS services if requested
    if [[ "$UPDATE_ECS" == "true" ]]; then
        for service in "${SERVICE_ARRAY[@]}"; do
            service=$(echo "$service" | xargs)
            update_ecs_service "$service"
        done
    fi

    print_summary

    log_info "Deployment complete!"
}

main "$@"
