# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Copy source code
COPY . .

# Build the application
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dudu-proxy .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/dudu-proxy .
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN addgroup -g 1000 dudu && \
    adduser -D -u 1000 -G dudu dudu && \
    chown -R dudu:dudu /app

USER dudu

# Expose ports
EXPOSE 8080 1080

# Run the application
ENTRYPOINT ["./dudu-proxy"]
CMD ["-config", "configs/config.example.json"]
