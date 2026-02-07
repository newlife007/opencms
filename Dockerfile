# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API server (using main.go in root directory)
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates ffmpeg

WORKDIR /app

# Copy binary and config
COPY --from=builder /build/api .
COPY --from=builder /build/configs ./configs

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
CMD ["./api"]
