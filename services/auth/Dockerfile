## --------------------------------------------------------------------
## Build stage
## --------------------------------------------------------------------
FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy Go workspace definitions and modules
COPY go.work go.work.sum ./
COPY libs/go/auth ./libs/go/auth
COPY services/auth ./services/auth

# Use the workspace while building
ENV GOWORK=/build/go.work

WORKDIR /build/services/auth

# Download dependencies (leverages go.work)
RUN go mod download

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/auth ./cmd/auth

## --------------------------------------------------------------------
## Runtime stage
## --------------------------------------------------------------------
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata wget \
    && addgroup -g 1000 appuser \
    && adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/auth .

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./auth"]

