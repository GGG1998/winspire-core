# syntax=docker/dockerfile:1.6

FROM golang:1.25-alpine3.21 AS builder
WORKDIR /src

# Pre-fetch dependencies before copying the full source to leverage Docker layer caching.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/game-management ./cmd/game-management

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app

# Copy binary
COPY --from=builder /out/game-management /usr/local/bin/game-management

# Copy migrations
COPY --from=builder /src/migrations ./migrations

ENV APP_ENV=production \
    SERVICE_PORT=8087

EXPOSE 8087

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8087/healthz || exit 1

USER app
ENTRYPOINT ["/usr/local/bin/game-management"]
