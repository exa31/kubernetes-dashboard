# Multi-stage Dockerfile for KubeNexus Go Backend
# Stage 1: Build binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Cache Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY . .

# Compile static binary for Linux with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Stage 2: Minimal, secure runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

# Run as non-root user for K8s security best practice
RUN adduser -D -u 1000 appuser

WORKDIR /app

COPY --from=builder /app/server .

USER appuser

EXPOSE 3001

ENTRYPOINT ["./server"]
