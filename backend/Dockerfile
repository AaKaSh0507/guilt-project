# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Allow toolchain download for newer Go versions
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the API binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /guiltmachine-backend ./cmd/api

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /guiltmachine-backend .

# Copy migrations for potential in-container migration runs
COPY --from=builder /app/migrations ./migrations

EXPOSE 9090

CMD ["./guiltmachine-backend"]
