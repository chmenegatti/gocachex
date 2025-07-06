# Multi-stage build for GoCacheX CLI
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the CLI application
RUN make build && go build -o /app/bin/gocachex-cli ./examples/cli

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 gocachex && \
    adduser -D -s /bin/sh -u 1000 -G gocachex gocachex

# Set working directory
WORKDIR /home/gocachex

# Copy binary from builder
COPY --from=builder /app/bin/gocachex-cli /usr/local/bin/gocachex-cli
COPY --from=builder /app/examples/configs/ /home/gocachex/configs/

# Set permissions
RUN chown -R gocachex:gocachex /home/gocachex

# Switch to non-root user
USER gocachex

# Expose default ports (if running as server)
EXPOSE 8080 9090

# Set default command
ENTRYPOINT ["gocachex-cli"]
CMD ["-help"]

# Labels for metadata
LABEL org.opencontainers.image.title="GoCacheX"
LABEL org.opencontainers.image.description="Distributed cache library for Go"
LABEL org.opencontainers.image.vendor="Cesar Menegatti"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/chmenegatti/gocachex"
LABEL org.opencontainers.image.documentation="https://github.com/chmenegatti/gocachex/blob/main/README.md"
