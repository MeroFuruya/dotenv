# Build stage
FROM golang:latest AS builder

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dotenv .

# Final stage
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy the binary from builder stage
COPY --from=builder /app/dotenv /dotenv

# Use distroless nonroot user
USER nonroot:nonroot

# Set the binary as entrypoint
ENTRYPOINT ["/dotenv"]