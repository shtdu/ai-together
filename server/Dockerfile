# Build stage
FROM golang:1.25-bookworm AS builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    make \
    && rm -rf /var/lib/apt/lists/*

# Copy dependency files first for better caching
COPY shared/go.mod shared/go.sum ./shared/
COPY server/go.mod server/go.sum ./server/

# Download dependencies (this layer will be cached unless go.mod changes)
RUN cd shared && go mod download && \
    cd ../server && go mod download

# Copy source code
COPY shared/ ./shared/
COPY server/ ./server/
COPY Makefile ./

# Build production binary optimized with -ldflags
RUN make server-production

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install ca-certificates and wget for HTTPS requests and health checks
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates wget && \
    rm -rf /var/lib/apt/lists/*

# Copy binary and migrations from builder
COPY --from=builder /build/server/bin/codetogether /app/
COPY --from=builder /build/server/migrations /app/migrations

# Expose port
EXPOSE 9080

# Run as non-root user for security
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9080/health || exit 1

# Run the server
CMD ["/app/codetogether"]
