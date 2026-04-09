# Build stage
FROM golang:1.26.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Runtime stage
FROM alpine:3.21

# Install only what's needed: ca-certificates for TLS, wget for healthcheck
RUN apk --no-cache add ca-certificates wget

# Create non-root group and user with /bin/sh for debugging
RUN addgroup -S appgroup && \
    adduser -S -G appgroup -s /bin/sh appuser

WORKDIR /app

# Copy binary and migrations from builder (with ownership set at copy time)
COPY --chown=appuser:appgroup --from=builder /app/main .
COPY --chown=appuser:appgroup --from=builder /app/migrations ./migrations

# Run as non-root
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz

CMD ["./main"]
