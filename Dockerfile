############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Create appuser.
RUN adduser -D -g '' appuser

# Set working directory.
WORKDIR /app

# Copy go.mod and go.sum first for dependency caching.
COPY go.mod go.sum ./

# Download dependencies.
RUN go mod download

# Copy the rest of the application source code.
COPY . ./

# Build the binary for the target platform.
# GOARCH and GOOS will be set automatically by Docker Buildx depending on the platform target.
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/dolr ./cmd/main.go

# Copy additional files required in the final image.
RUN cp dolr.example.json /app/dolr.json

############################
# STEP 2 build a small image
############################
FROM alpine

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd

# Copy the built application and necessary files from the builder.
COPY --from=builder /app /app

WORKDIR /app

# Set file permissions to appuser.
RUN chown -R appuser .

# Use an unprivileged user for security.
USER appuser

# Expose the necessary port.
EXPOSE 8080

# Run the dolr binary.
ENTRYPOINT ["/app/dolr"]
