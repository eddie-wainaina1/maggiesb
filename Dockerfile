# Build stage
FROM golang:1.25.6-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/maggiesb .

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/bin/maggiesb .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./maggiesb"]
