# Build stage
FROM golang:1.23-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o /app/bin/grpc-server ./cmd/grpc/main.go

# Final stage
FROM alpine:3.19

# Add ca certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create necessary directories
RUN mkdir -p /app/config

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/grpc-server .
COPY --from=builder /app/config/config.yaml ./config/

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 50051

# Set environment variables
ENV APP_ENV=production

# Command to run
CMD ["./grpc-server"]
