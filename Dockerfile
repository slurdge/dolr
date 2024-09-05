############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Create appuser.
RUN adduser -D -g '' appuser

WORKDIR $GOPATH/src/dolr/
COPY src/dolr .

# Fetch dependencies.

# Using go get.
WORKDIR $GOPATH/src/dolr/cmd
RUN go get -d -v

# Using go mod.
# RUN go mod download

# Build the binary.
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/dolr
RUN cp dolr.example.json /go/bin/dolr.json
RUN cp -r static templates /go/bin/

############################
# STEP 2 build a small image
############################
FROM alpine

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable.
COPY --from=builder /go/bin /go/bin

WORKDIR /go/bin/

RUN chown -R appuser .

# Use an unprivileged user.
USER appuser

EXPOSE 8080

# Run the hello binary.
ENTRYPOINT ["/go/bin/dolr"]
