# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install build dependencies (gcc needed for cgo used by some drivers)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependency downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build a statically linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server ./cmd/api

# Stage 2: Runtime
FROM alpine:3.20

# ca-certificates: needed for outbound TLS (Resend API, etc.)
# tzdata: needed if the app uses time zones
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

# The app reads config from environment variables; no .env file is baked in
EXPOSE 8000

ENTRYPOINT ["/app/server"]
