# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application (parameterized with ARG)
ARG SERVICE=api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/${SERVICE}

# Final stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Shanghai

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/app .

# Copy configuration files
ARG SERVICE=api
ARG CONFIG_ENV=prod
COPY --from=builder /app/config/${CONFIG_ENV}/.env /app/config/${CONFIG_ENV}/.env

# Create a non-root user and set permissions
RUN adduser -D -H -h /app appuser && \
    chown -R appuser:appuser /app
USER appuser

# Expose the port
EXPOSE 8080

# Run the application with environment config
ENV GO_ENV=${CONFIG_ENV}
CMD ["./app"]
