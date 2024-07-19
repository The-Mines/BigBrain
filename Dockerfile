# Build stage
FROM --platform=$BUILDPLATFORM golang:latest AS builder

WORKDIR /app

# Copy go mod files
COPY src/go.mod src/go.sum ./
COPY src/ .
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
        "linux/amd64")  GOOS=linux   GOARCH=amd64 ;; \
        "linux/arm64")  GOOS=linux   GOARCH=arm64 ;; \
        "darwin/amd64") GOOS=darwin  GOARCH=amd64 ;; \
        "darwin/arm64") GOOS=darwin  GOARCH=arm64 ;; \
        *)              GOOS=linux   GOARCH=amd64 ;; \
    esac && \
    echo "Building for $GOOS/$GOARCH" && \
    GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o BigBrain .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/BigBrain .

# Set the entrypoint to the BigBrain binary
ENTRYPOINT ["./BigBrain"]

# Set default command (can be overridden)
CMD ["-h"]