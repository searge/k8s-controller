# syntax=docker/dockerfile:1.4
FROM golang:1.25-alpine AS builder

# Install ca-certificates and git for dependency downloads
RUN apk add --no-cache ca-certificates git

# Create non-root user for build process
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Download dependencies (cached layer if go.mod/go.sum haven't changed)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments for cross-compilation and versioning
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev

# Build the application with optimizations
RUN CGO_ENABLED=0 \
    GOOS="${TARGETOS}" \
    GOARCH="${TARGETARCH}" \
    go build \
    -ldflags="-w -s -X=github.com/Searge/k8s-controller/cmd.Version=${VERSION}" \
    -o kc \
    main.go

# Final stage - minimal runtime image
FROM gcr.io/distroless/static-debian12:nonroot

# Metadata labels following OCI Image Format Specification
LABEL org.opencontainers.image.title="k8s-controller" \
      org.opencontainers.image.description="A production-grade Golang Kubernetes controller" \
      org.opencontainers.image.url="https://github.com/Searge/k8s-controller" \
      org.opencontainers.image.source="https://github.com/Searge/k8s-controller" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.vendor="Sergij Boremchuk" \
      org.opencontainers.image.licenses="GPL-3.0-or-later"

# Copy the binary from builder stage
COPY --from=builder /app/kc /kc

# Use non-root user (already provided by distroless:nonroot)
USER nonroot:nonroot

# Expose default port
EXPOSE 8080

# Health check to ensure container is healthy
# Note: HEALTHCHECK is Docker-specific and may not work with all OCI runtimes
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/kc", "version"]

# Default command
ENTRYPOINT ["/kc"]
CMD ["serve"]
