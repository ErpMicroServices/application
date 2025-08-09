# GraphQL Federation Gateway (Go) - Multi-stage Docker build

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run quality gates and build
RUN make quality-gates build-linux

# Final stage - minimal runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S gateway && \
    adduser -u 1001 -S gateway -G gateway

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/graphql-federation-gateway_unix /app/graphql-federation-gateway

# Copy configuration files
COPY --from=builder /app/.env.gateway.example /app/.env.example

# Set ownership
RUN chown -R gateway:gateway /app

# Switch to non-root user
USER gateway

# Expose ports
EXPOSE 4000 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4000/health || exit 1

# Set environment variables
ENV ENVIRONMENT=production
ENV SERVER_PORT=4000
ENV METRICS_PORT=9090

# Run the application
CMD ["./graphql-federation-gateway"]