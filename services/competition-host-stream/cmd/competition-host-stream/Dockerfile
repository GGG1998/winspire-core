# syntax=docker/dockerfile:1.6

FROM golang:1.22-alpine AS builder
WORKDIR /src

# Pre-fetch dependencies before copying the full source to leverage Docker layer caching.
COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/competition-host-stream ./cmd/competition-host-stream

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/competition-host-stream /usr/local/bin/competition-host-stream

ENV APP_ENV=production \
    SERVICE_PORT=8086
EXPOSE 8086

USER app
ENTRYPOINT ["/usr/local/bin/competition-host-stream"]

